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
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/app"
	"github.com/kkevinchou/izzet/app/apputils"
	"github.com/kkevinchou/izzet/app/client/edithistory"
	"github.com/kkevinchou/izzet/app/client/editorcamera"
	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/izzet/app/render"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/app/server"
	"github.com/kkevinchou/izzet/app/systems/clientsystems"
	"github.com/kkevinchou/izzet/internal/assets"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/contentbrowser"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
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

func (g *Client) Platform() platforms.Platform {
	return g.platform
}

func (g *Client) saveWorld(worldFilePath string) {
	err := serialization.WriteToFile(g.world, worldFilePath)
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

	g.SelectEntity(nil)
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
	g.SelectEntity(nil)
	g.SetWorld(liveWorld)
}

func (g *Client) StopLiveWorld() {
	if g.AppMode() != app.AppModePlay {
		return
	}
	g.appMode = app.AppModeEditor
	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	g.SelectEntity(nil)
	g.SetWorld(g.editorWorld)
}

func (g *Client) AppMode() app.AppMode {
	return g.appMode
}

func (g *Client) CollisionObserver() *collisionobserver.CollisionObserver {
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

	g.ConfigureUI(false)

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
		g.ConfigureUI(true)
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
	globals.SetClientMetricsRegistry(g.metricsRegistry)
	g.collisionObserver = collisionobserver.NewCollisionObserver()
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

// registerSingleEntity registers the asset found at assetFilePath with the model library and asset manager
func (g *Client) registerSingleEntity(assetFilePath string) bool {
	baseFileName := apputils.NameFromAssetFilePath(assetFilePath)
	if g.AssetManager().LoadDocument(baseFileName, assetFilePath) {
		document := g.AssetManager().GetDocument(baseFileName)
		g.ModelLibrary().RegisterSingleEntityDocument(document)
		return true
	}
	return false
}

// TODO - import props? single vs multiple entities, animation, material, etc
// ImportToContentBrowser registers the asset found at assetFilePath with the model library and asset manager
// then, the asset is registered with the content browser
func (g *Client) ImportToContentBrowser(assetFilePath string) {
	if g.registerSingleEntity(assetFilePath) {
		baseFileName := apputils.NameFromAssetFilePath(assetFilePath)
		document := g.AssetManager().GetDocument(baseFileName)
		g.contentBrowser.AddGLTFModel(assetFilePath, document)

		var primitiveSpecs []*modelspec.PrimitiveSpecification
		for _, mesh := range document.Meshes {
			primitiveSpecs = append(primitiveSpecs, mesh.Primitives...)

		}
	}
}

func (g *Client) LoadProject(name string) bool {
	if name == "" {
		return false
	}

	f, err := os.Open(apputils.PathToProjectFile(name))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var project Project
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&project)
	if err != nil {
		panic(err)
	}

	g.projectName = project.Name
	g.contentBrowser = &project.ContentBrowser

	for _, item := range g.contentBrowser.Items {
		g.registerSingleEntity(item.InFilePath)
	}

	return g.loadWorld(path.Join(settings.ProjectsDirectory, name, name+".json"))
}

func (g *Client) SaveProject() {
	err := os.MkdirAll(filepath.Join(settings.ProjectsDirectory, g.projectName), os.ModePerm)
	if err != nil {
		panic(err)
	}

	worldFilePath := path.Join(settings.ProjectsDirectory, g.projectName, fmt.Sprintf("./%s.json", g.projectName))
	g.saveWorld(worldFilePath)

	err = os.MkdirAll(filepath.Join(settings.ProjectsDirectory, g.projectName, "content"), os.ModePerm)
	if err != nil {
		panic(err)
	}

	items := g.contentBrowser.Items
	for i := range items {
		// this has already been saved, skip
		if items[i].SavedToProjectFolder {
			continue
		}

		baseFileName := strings.Split(filepath.Base(items[i].InFilePath), ".")[0]
		parentDirectory := filepath.Dir(items[i].InFilePath)

		var fileNames []string
		fileNames = append(fileNames, baseFileName+filepath.Ext(items[i].InFilePath))
		for _, fileName := range items[i].PeripheralFiles {
			fileNames = append(fileNames, fileName)
		}

		for _, fileName := range fileNames {
			importedFile, err := os.Open(filepath.Join(parentDirectory, fileName))
			if err != nil {
				panic(err)
			}
			defer importedFile.Close()

			fileBytes, err := io.ReadAll(importedFile)
			if err != nil {
				panic(err)
			}

			outFilePath := filepath.Join(settings.ProjectsDirectory, g.projectName, "content", fileName)
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

		// overwrite in file path to be the asset copy in in the project folder
		items[i].SavedToProjectFolder = true
		items[i].InFilePath = filepath.Join(settings.ProjectsDirectory, g.projectName, "content", baseFileName+filepath.Ext(items[i].InFilePath))
	}

	f, err := os.OpenFile(filepath.Join(settings.ProjectsDirectory, g.projectName, "main_project.izt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	project := Project{
		Name:           g.projectName,
		WorldFile:      worldFilePath,
		ContentBrowser: *g.contentBrowser,
	}

	encoder := json.NewEncoder(f)
	encoder.Encode(project)
}

func (g *Client) SaveProjectAs(name string) {
	g.projectName = name
	g.SaveProject()
}

func (g *Client) Shutdown() {
	g.gameOver = true
}

func (g *Client) ConfigureUI(enabled bool) {
	g.runtimeConfig.UIEnabled = enabled
	g.renderer.ConfigureUI()
}

func (g *Client) SetWindowSize(width, height int) {
	g.width, g.height = width, height
}

func (g *Client) WindowSize() (int, int) {
	return g.width, g.height
}

func (g *Client) Minimized() bool {
	return g.window.Minimized()
}

func (g *Client) WindowFocused() bool {
	return g.window.WindowFocused()
}

func (g *Client) ContentBrowser() *contentbrowser.ContentBrowser {
	return g.contentBrowser
}

func (g *Client) SelectEntity(entity *entities.Entity) {
	g.selectedEntity = entity
}

func (g *Client) SelectedEntity() *entities.Entity {
	return g.selectedEntity
}

func (g *Client) InstantiateEntity(entityHandle string) *entities.Entity {
	document := g.AssetManager().GetDocument(entityHandle)
	handle := modellibrary.NewGlobalHandle(entityHandle)
	if len(document.Scenes) != 1 {
		panic("single entity asset loading only supports a singular scene")
	}

	scene := document.Scenes[0]
	node := scene.Nodes[0]

	entity := entities.InstantiateEntity(entityHandle)
	entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Ident4(), Visible: true, ShadowCasting: true}
	var vertices []modelspec.Vertex
	entities.VerticesFromNode(node, document, &vertices)
	entity.InternalBoundingBox = collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
	entities.SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
	entities.SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
	entities.SetScale(entity, utils.Vec3F32ToF64(node.Scale))
	// entities.SetScale(entity, mgl64.Vec3{4, 4, 4})

	primitives := g.ModelLibrary().GetPrimitives(handle)
	if len(primitives) > 0 {
		entity.Collider = &entities.ColliderComponent{
			ColliderGroup:   entities.ColliderGroupMap[entities.ColliderGroupTerrain],
			TriMeshCollider: collider.CreateTriMeshFromPrimitives(entities.MLPrimitivesTospecPrimitive(primitives)),
		}
	}

	g.world.AddEntity(entity)
	return entity
}

func (g *Client) BuildNavMesh(app renderiface.App, world renderiface.GameWorld, iterationCount int, walkableHeight int, climbableHeight int, minRegionArea int, maxError float64) {
	start := time.Now()
	defer func() {
		fmt.Println("BuildNavMesh completed in", time.Since(start))
	}()
	minVertex := mgl64.Vec3{-500, -250, -500}
	maxVertex := mgl64.Vec3{500, 250, 500}

	vxs := int(maxVertex.X() - minVertex.X())
	vzs := int(maxVertex.Z() - minVertex.Z())
	nmbb := collider.BoundingBox{MinVertex: minVertex, MaxVertex: maxVertex}

	hf := navmesh.NewHeightField(vxs, vzs, minVertex, maxVertex)
	var debugLines [][2]mgl64.Vec3

	for _, entity := range world.Entities() {
		if entity.MeshComponent == nil {
			continue
		}
		if !entity.HasBoundingBox() {
			continue
		}

		ebb := entity.BoundingBox()

		if ebb.MaxVertex.X() < nmbb.MinVertex.X() || ebb.MinVertex.X() > nmbb.MaxVertex.X() {
			continue
		}
		if ebb.MaxVertex.Y() < nmbb.MinVertex.Y() || ebb.MinVertex.Y() > nmbb.MaxVertex.Y() {
			continue
		}
		if ebb.MaxVertex.Z() < nmbb.MinVertex.Z() || ebb.MinVertex.Z() > nmbb.MaxVertex.Z() {
			continue
		}

		primitives := app.ModelLibrary().GetPrimitives(entity.MeshComponent.MeshHandle)
		transform := utils.Mat4F64ToF32(entities.WorldTransform(entity))
		up := mgl64.Vec3{0, 1, 0}
		right := mgl64.Vec3{1, 0, 0}

		// rasterize triangles
		for _, p := range primitives {
			for i := 0; i < len(p.Primitive.Vertices); i += 3 {
				v1 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i].Position.Vec4(1)).Vec3())
				v2 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+1].Position.Vec4(1)).Vec3())
				v3 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+2].Position.Vec4(1)).Vec3())

				// debugLines = append(debugLines, [2]mgl64.Vec3{v1, v2})
				// debugLines = append(debugLines, [2]mgl64.Vec3{v2, v3})
				// debugLines = append(debugLines, [2]mgl64.Vec3{v3, v1})

				tv1 := v2.Sub(v1)
				tv2 := v3.Sub(v2)

				normal := tv1.Cross(tv2)
				if normal.LenSqr() > 0 {
					normal = normal.Normalize()
				}
				isUp := normal.Dot(up) > 0.7
				isDown := normal.Dot(up) < -0.7
				isRight := normal.Dot(right) > 0.7
				isLeft := normal.Dot(right) < -0.7
				_, _, _, _ = isUp, isDown, isRight, isLeft

				// walkable := isUp
				// if (isUp && entity.GetID() < 3) || ((isDown || isUp) && entity.GetID() >= 3) {
				// if (isUp && entity.GetID() < 3) || (isUp && entity.GetID() >= 3) {
				// if (isUp && entity.GetID() < 3) || (isDown && entity.GetID() >= 3) {
				if isUp {
					navmesh.RasterizeTriangle(
						int(v1.X()),
						int(v1.Y()),
						int(v1.Z()),
						int(v2.X()),
						int(v2.Y()),
						int(v2.Z()),
						int(v3.X()),
						int(v3.Y()),
						int(v3.Z()),
						hf,
						isUp,
					)
				}
				// }
			}
		}
	}

	hf.Test()

	navmesh.FilterLowHeightSpans(walkableHeight, hf)
	chf := navmesh.NewCompactHeightField(walkableHeight, climbableHeight, hf)
	chf.Test()
	navmesh.BuildDistanceField(chf)

	navmesh.BuildRegions(chf, iterationCount, minRegionArea, 1)
	contourSet := navmesh.BuildContours(chf, maxError, 1)

	for _, contour := range contourSet.Contours {
		verts := contour.Verts
		for i := 0; i < len(verts); i++ {
			v1 := verts[i]
			v2 := verts[(i+1)%len(verts)]
			v164 := mgl64.Vec3{float64(v1.X), float64(v1.Y + 2), float64(v1.Z)}.Add(minVertex)
			v264 := mgl64.Vec3{float64(v2.X), float64(v2.Y + 2), float64(v2.Z)}.Add(minVertex)

			debugLines = append(debugLines, [2]mgl64.Vec3{v164, v264})
		}
	}

	// lines don't join properly at mgl64.Vec3{267, 173, 129}
	// debugLines = [][2]mgl64.Vec3{
	// 	// [2]mgl64.Vec3{mgl64.Vec3{266, 173, -28}, mgl64.Vec3{267, 173, 129}},
	// 	[2]mgl64.Vec3{mgl64.Vec3{267, 173, 129}, mgl64.Vec3{266, 173, -28}},
	// 	[2]mgl64.Vec3{mgl64.Vec3{267, 173, 129}, mgl64.Vec3{209, 173, 302}},
	// }

	nm := &navmesh.NavigationMesh{
		HeightField:          hf,
		CompactHeightField:   chf,
		Volume:               nmbb,
		BlurredDistances:     chf.Distances(),
		DebugLines:           debugLines,
		Invalidated:          true,
		InvalidatedTimestamp: time.Now().Second(),
	}

	g.navMesh = nm
}

func (g *Client) NavMesh() *navmesh.NavigationMesh {
	return g.navMesh
}
