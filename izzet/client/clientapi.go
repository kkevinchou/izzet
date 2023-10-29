package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/systems/clientsystems/commandframe"
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

func (g *Client) Camera() *camera.Camera {
	return g.camera
}

func (g *Client) Platform() *input.SDLPlatform {
	return g.platform
}

func (g *Client) Serializer() *serialization.Serializer {
	return g.serializer
}

func (g *Client) SaveWorld(name string) {
	g.serializer.WriteToFile(g.world, fmt.Sprintf("./%s.json", name))
}

func (g *Client) LoadWorld(name string) bool {
	if name == "" {
		return false
	}

	filename := fmt.Sprintf("./%s.json", name)
	world, err := g.serializer.ReadFromFile(filename)
	if err != nil {
		fmt.Println("failed to load world", filename, err)
		panic(err)
	}

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

func (g *Client) NavMesh() *navmesh.NavigationMesh {
	return g.navigationMesh
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
	err := g.serializer.Write(g.world, &buffer)
	if err != nil {
		panic(err)
	}

	liveWorld, err := g.serializer.Read(&buffer)
	if err != nil {
		panic(err)
	}

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

// computes the near plane position for a given x y coordinate
func (g *Client) NDCToWorldPosition(viewerContext render.ViewerContext, directionVec mgl64.Vec3) mgl64.Vec3 {
	// ndcP := mgl64.Vec4{((x / float64(g.width)) - 0.5) * 2, ((y / float64(g.height)) - 0.5) * -2, -1, 1}
	nearPlanePos := viewerContext.InverseViewMatrix.Inv().Mul4(viewerContext.ProjectionMatrix.Inv()).Mul4x1(directionVec.Vec4(1))
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}

func (g *Client) WorldToNDCPosition(viewerContext render.ViewerContext, worldPosition mgl64.Vec3) (mgl64.Vec2, bool) {
	screenPos := viewerContext.ProjectionMatrix.Mul4(viewerContext.InverseViewMatrix).Mul4x1(worldPosition.Vec4(1))
	behind := screenPos.Z() < 0
	screenPos = screenPos.Mul(1 / screenPos.W())
	return screenPos.Vec2(), behind
}

func (g *Client) PhysicsObserver() *observers.PhysicsObserver {
	return g.physicsObserver
}

func (g *Client) Settings() *app.Settings {
	return g.settings
}

func (g *Client) Connect() {
	if g.IsConnected() {
		return
	}

	g.StartLiveWorld()

	address := fmt.Sprintf("localhost:7878")
	fmt.Println("connecting to " + address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	g.client = network.NewClient(conn)
	messageTransport, err := g.client.Recv()
	if err != nil {
		panic(err)
	}

	message, err := network.ExtractMessage[network.AckPlayerJoinMessage](messageTransport)
	if err != nil {
		panic(err)
	}

	g.playerID = message.PlayerID
	g.connection = conn
	g.networkMessages = make(chan network.MessageTransport, 100)

	// initialize the player's camera and entity
	var entity entities.Entity
	err = json.Unmarshal(message.EntityBytes, &entity)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
	}
	serialization.InitDeserializedEntity(&entity, g.ModelLibrary(), false)
	g.world.AddEntity(&entity)

	var camera entities.Entity
	err = json.Unmarshal(message.CameraBytes, &camera)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
	}
	serialization.InitDeserializedEntity(&camera, g.ModelLibrary(), false)
	g.world.AddEntity(&camera)

	g.SetPlayerCamera(&camera)
	g.SetPlayerEntity(&entity)

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
	g.connected = true
	fmt.Println("finished connect")
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
	return g.connected
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

func (g *Client) GetCommandFrameHistory() *commandframe.CommandFrameHistory {
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
