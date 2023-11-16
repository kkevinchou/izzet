package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/client/editorcamera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/project"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems/clientsystems"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

func (g *Client) GetPrefabByID(id int) *prefabs.Prefab {
	return g.prefabs[id]
}

func (g *Client) Prefabs() []*prefabs.Prefab {
	var ids []int
	for id, _ := range g.prefabs {
		ids = append(ids, id)
	}

	sort.Ints(ids)

	ps := []*prefabs.Prefab{}
	for _, id := range ids {
		ps = append(ps, g.prefabs[id])
	}

	return ps
}

func (g *Client) AssetManager() *assets.AssetManager {
	return g.assetManager
}

func (g *Client) ModelLibrary() *modellibrary.ModelLibrary {
	return g.modelLibrary
}

func (g *Client) GetEditorCameraPosition() mgl64.Vec3 {
	return g.camera.Position
}

func (g *Client) GetEditorCameraRotation() mgl64.Quat {
	return g.camera.Rotation
}

func (g *Client) Platform() *input.SDLPlatform {
	return g.platform
}

func (g *Client) saveWorld(name string) {
	err := serialization.WriteToFile(g.world, path.Join(settings.ProjectsDirectory, name, fmt.Sprintf("./%s.json", name)))
	if err != nil {
		panic(err)
	}
}

func (g *Client) loadWorld(filepath string) bool {
	if filepath == "" {
		return false
	}

	world, err := serialization.ReadFromFile(filepath)
	if err != nil {
		fmt.Println("failed to load world", filepath, err)
		panic(err)
	}
	serialization.InitDeserializedEntities(world.Entities(), g.modelLibrary)

	g.editHistory.Clear()
	g.world.SpatialPartition().Clear()

	var maxID int
	for _, e := range world.Entities() {
		if e.ID > maxID {
			maxID = e.ID
		}
	}
	entities.SetNextID(maxID + 1)

	panels.SelectEntity(nil)
	g.SetWorld(world)
	return true
}

// game world
func (g *Client) AppendEdit(edit edithistory.Edit) {
	g.editHistory.Append(edit)
}

// game world
func (g *Client) Redo() {
	g.editHistory.Redo()
}

// game world
func (g *Client) Undo() {
	g.editHistory.Undo()
}

func (g *Client) ResetNavMeshVAO() {
	render.ResetNavMeshVAO = true
}

func (g *Client) SetShowImguiDemo(value bool) {
	g.showImguiDemo = value
}

func (g *Client) ShowImguiDemo() bool {
	return g.showImguiDemo
}

func (g *Client) MetricsRegistry() *metrics.MetricsRegistry {
	return g.metricsRegistry
}

func (g *Client) SetWorld(world *world.GameWorld) {
	g.world = world
	g.renderer.SetWorld(world)
}

func (g *Client) StartLiveWorld() {
	if g.AppMode() != app.AppModeEditor {
		return
	}
	g.appMode = app.AppModePlay
	g.editorWorld = g.world

	var buffer bytes.Buffer
	err := serialization.Write(g.world, &buffer)
	if err != nil {
		panic(err)
	}

	liveWorld, err := serialization.Read(&buffer)
	if err != nil {
		panic(err)
	}
	serialization.InitDeserializedEntities(liveWorld.Entities(), g.modelLibrary)

	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	panels.SelectEntity(nil)
	g.SetWorld(liveWorld)
}

func (g *Client) StopLiveWorld() {
	if g.AppMode() != app.AppModePlay {
		return
	}
	g.appMode = app.AppModeEditor
	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	panels.SelectEntity(nil)
	g.SetWorld(g.editorWorld)
}

func (g *Client) AppMode() app.AppMode {
	return g.appMode
}

func (g *Client) CollisionObserver() *observers.CollisionObserver {
	return g.collisionObserver
}

func (g *Client) RuntimeConfig() *app.RuntimeConfig {
	return g.runtimeConfig
}
func (g *Client) ConnectAndInitialize() error {
	g.initialize()
	return g.Connect()
}

func (g *Client) Connect() error {
	if g.IsConnected() {
		return nil
	}

	fmt.Println("connecting to " + g.serverAddress)

	conn, err := net.Dial("tcp", g.serverAddress)
	if err != nil {
		return err
	}

	g.runtimeConfig.UIEnabled = false

	g.StartLiveWorld()

	g.client = network.NewClient(conn)
	messageTransport, err := g.client.Recv()
	if err != nil {
		return err
	}

	message, err := network.ExtractMessage[network.AckPlayerJoinMessage](messageTransport)
	if err != nil {
		return err
	}

	g.playerID = message.PlayerID
	g.connection = conn
	g.networkMessages = make(chan network.MessageTransport, 100)

	// initialize the player's camera and playerEntity
	var playerEntity entities.Entity
	err = json.Unmarshal(message.EntityBytes, &playerEntity)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
	}
	serialization.InitDeserializedEntity(&playerEntity, g.ModelLibrary())
	g.world.AddEntity(&playerEntity)

	var camera entities.Entity
	err = json.Unmarshal(message.CameraBytes, &camera)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
	}
	serialization.InitDeserializedEntity(&camera, g.ModelLibrary())
	g.world.AddEntity(&camera)

	g.SetPlayerCamera(&camera)
	g.SetPlayerEntity(&playerEntity)

	fmt.Println("CLIENT player id", playerEntity.GetID(), "camera id", camera.GetID())

	world, err := serialization.Read(bytes.NewReader(message.Snapshot))
	if err != nil {
		return err
	}
	serialization.InitDeserializedEntities(world.Entities(), g.modelLibrary)

	for _, entity := range world.Entities() {
		if entity.GetID() == camera.GetID() || entity.GetID() == playerEntity.GetID() {
			continue
		}
		g.world.AddEntity(entity)
	}

	// TODO a done channel to close out the goroutine
	go func() {
		defer conn.Close()

		for {
			message, err := g.client.Recv()
			if err != nil {
				if err == io.EOF {
					continue
				}

				fmt.Println("error reading incoming message:", err.Error())
				fmt.Println("closing connection")
				return
			}

			g.networkMessages <- message
		}
	}()
	g.clientConnected = true
	fmt.Println("finished connect")
	return nil
}

func (g *Client) NetworkMessagesChannel() chan network.MessageTransport {
	return g.networkMessages
}

func (g *Client) CommandFrame() int {
	return g.commandFrame
}

func (g *Client) GetPlayerID() int {
	return g.playerID

}
func (g *Client) IsConnected() bool {
	return g.clientConnected
}
func (g *Client) GetPlayerConnection() net.Conn {
	return g.connection
}
func (g *Client) SetPlayerEntity(entity *entities.Entity) {
	g.playerEntity = entity
}
func (g *Client) SetPlayerCamera(entity *entities.Entity) {
	g.playerCamera = entity
}
func (g *Client) GetPlayerEntity() *entities.Entity {
	return g.playerEntity
}
func (g *Client) GetPlayerCamera() *entities.Entity {
	return g.playerCamera
}

func (g *Client) GetCommandFrameHistory() *clientsystems.CommandFrameHistory {
	return g.commandFrameHistory
}

func (g *Client) Client() network.IzzetClient {
	return g.client
}

func (g *Client) IsServer() bool {
	return false
}

func (g *Client) IsClient() bool {
	return true
}

func (g *Client) GetPlayer(playerID int) *network.Player {
	panic("wat")
}

func (g *Client) StartAsyncServer() {
	started := make(chan bool)

	go func() {
		var worldBytes bytes.Buffer
		err := serialization.Write(g.world, &worldBytes)
		if err != nil {
			panic(err)
		}

		world, err := serialization.Read(&worldBytes)
		if err != nil {
			panic(err)
		}
		serialization.InitDeserializedEntities(world.Entities(), g.modelLibrary)

		serverApp := server.NewWithWorld("_assets", world)
		serverApp.Start(started, g.asyncServerDone)
		g.asyncServerStarted = false
		fmt.Println("Server finished teardown")
	}()

	<-started
	g.asyncServerStarted = true
}

func (g *Client) DisconnectAsyncServer() {
	g.DisconnectClient()
	g.asyncServerDone <- true
}

func (g *Client) AsyncServerStarted() bool {
	return g.asyncServerStarted
}

func (g *Client) DisconnectClient() {
	if g.clientConnected {
		g.connection.Close()
		g.clientConnected = false
		g.commandFrameHistory.Reset()
		g.StopLiveWorld()
		g.runtimeConfig.UIEnabled = true
	}
}

func (g *Client) World() *world.GameWorld {
	return g.world
}

func (g *Client) StateBuffer() *clientsystems.StateBuffer {
	return g.stateBuffer
}

func (g *Client) initialize() {
	g.stateBuffer = clientsystems.NewStateBuffer()
	g.commandFrameHistory = clientsystems.NewCommandFrameHistory()

	g.camera = &editorcamera.Camera{
		Position: mgl64.Vec3{-82, 230, 95},
		Rotation: mgl64.QuatIdent(),
	}

	g.editHistory = edithistory.New()
	g.metricsRegistry = metrics.New()
	g.collisionObserver = observers.NewCollisionObserver()
	g.stateBuffer = clientsystems.NewStateBuffer()
}

func (g *Client) ServerAddress() string {
	return g.serverAddress
}

func (g *Client) GetFrameInput() input.Input {
	return g.frameInput
}

func (g *Client) GetFrameInputPtr() *input.Input {
	return &g.frameInput
}

func (g *Client) SetServerStats(stats serverstats.ServerStats) {
	g.serverStats = stats
}

func (g *Client) GetServerStats() serverstats.ServerStats {
	return g.serverStats
}

func (g *Client) GetProject() *project.Project {
	return g.project
}

func (g *Client) LoadProject(name string) bool {
	g.project.Name = name
	return g.loadWorld(path.Join(settings.ProjectsDirectory, name, name+".json"))
}

func (g *Client) SaveProject() {
	err := os.MkdirAll(filepath.Join(settings.ProjectsDirectory, g.project.Name), os.ModePerm)
	if err != nil {
		panic(err)
	}
	g.saveWorld(g.project.Name)

	err = os.MkdirAll(filepath.Join(settings.ProjectsDirectory, g.project.Name, "content"), os.ModePerm)
	if err != nil {
		panic(err)
	}

	for i := range g.project.Content {
		content := &g.project.Content[i]
		baseFileName := strings.Split(filepath.Base(content.InFilePath), ".")[0]

		importedFile, err := os.Open(content.InFilePath)
		if err != nil {
			panic(err)
		}
		defer importedFile.Close()

		fileBytes, err := io.ReadAll(importedFile)
		if err != nil {
			panic(err)
		}

		outFilePath := path.Join(settings.ProjectsDirectory, g.project.Name, "content", baseFileName+filepath.Ext(content.InFilePath))
		content.OutFilepath = outFilePath

		outFile, err := os.OpenFile(outFilePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer outFile.Close()

		_, err = outFile.Write(fileBytes)
		if err != nil {
			panic(err)
		}
	}

	f, err := os.OpenFile(filepath.Join(settings.ProjectsDirectory, g.project.Name, "main_project.izt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.Encode(g.project)
}

func (g *Client) SaveProjectAs(name string) {
	g.project.Name = name
	g.SaveProject()
}
