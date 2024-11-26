package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sort"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/client/edithistory"
	"github.com/kkevinchou/izzet/izzet/client/editorcamera"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/contentbrowser"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/material"
	"github.com/kkevinchou/izzet/izzet/materialbrowser"
	"github.com/kkevinchou/izzet/izzet/mode"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems/clientsystems"
	"github.com/kkevinchou/izzet/izzet/types"
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
	serialization.InitDeserializedEntities(world.Entities(), g.assetManager)

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

func (g *Client) MetricsRegistry() *metrics.MetricsRegistry {
	return g.metricsRegistry
}

func (g *Client) SetWorld(world *world.GameWorld) {
	g.world = world
	g.renderSystem.SetWorld(world)
}

func (g *Client) StartLiveWorld() {
	if g.AppMode() != mode.AppModeEditor {
		return
	}
	g.appMode = mode.AppModePlay
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
	serialization.InitDeserializedEntities(liveWorld.Entities(), g.assetManager)

	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	g.SelectEntity(nil)
	g.SetWorld(liveWorld)
}

func (g *Client) StopLiveWorld() {
	if g.AppMode() != mode.AppModePlay {
		return
	}
	g.appMode = mode.AppModeEditor
	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	g.SelectEntity(nil)
	g.SetWorld(g.editorWorld)
}

func (g *Client) AppMode() mode.AppMode {
	return g.appMode
}

func (g *Client) CollisionObserver() *collisionobserver.CollisionObserver {
	return g.collisionObserver
}

func (g *Client) RuntimeConfig() *runtimeconfig.RuntimeConfig {
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
	serialization.InitDeserializedEntity(&playerEntity, g.AssetManager())
	g.world.AddEntity(&playerEntity)

	var camera entities.Entity
	err = json.Unmarshal(message.CameraBytes, &camera)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
	}
	serialization.InitDeserializedEntity(&camera, g.AssetManager())
	g.world.AddEntity(&camera)

	g.SetPlayerCamera(&camera)
	g.SetPlayerEntity(&playerEntity)

	fmt.Println("CLIENT player id", playerEntity.GetID(), "camera id", camera.GetID())

	world, err := serialization.Read(bytes.NewReader(message.Snapshot))
	if err != nil {
		return err
	}
	serialization.InitDeserializedEntities(world.Entities(), g.assetManager)

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

	var compiledNavMesh *navmesh.CompiledNavMesh
	if g.navMesh != nil {
		compiledNavMesh = navmesh.CompileNavMesh(g.navMesh)
	}

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
		serialization.InitDeserializedEntities(world.Entities(), g.assetManager)

		serverApp := server.NewWithWorld("_assets", world)
		serverApp.CopyLoadedAnimations(
			g.assetManager.Animations,
			g.assetManager.Joints,
			g.assetManager.RootJoints,
		)

		serverApp.SetNavMesh(compiledNavMesh)
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
		Position: settings.EditorCameraStartPosition,
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

func (g *Client) ImportAsset(config assets.AssetConfig) {
	g.assetManager.LoadAndRegisterDocument(config)
}

func (g *Client) Shutdown() {
	g.gameOver = true
}

func (g *Client) ConfigureUI(enabled bool) {
	g.runtimeConfig.UIEnabled = enabled
	g.renderSystem.ConfigureUI()
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
	return g.project.ContentBrowser
}

func (g *Client) SelectEntity(entity *entities.Entity) {
	g.selectedEntity = entity
}

func (g *Client) SelectedEntity() *entities.Entity {
	return g.selectedEntity
}

// func CreateEntitiesFromDocument(document *modelspec.Document, ml *assets.AssetManager) []*Entity {
// 	var spawnedEntities []*Entity

// 	entityAsset := data.EntityAssets[document.Name]

// 	if entityAsset.SingleEntity {
// 		handle := assets.NewSingleMeshHandle(document.Name)
// 		// entity := InstantiateEntity(document.Name)
// 		// entity.MeshComponent = &MeshC
// 		var scene *modelspec.Scene
// 		if len(document.Scenes) != 1 {
// 			panic("single entity asset loading only supports a singular scene")
// 		}
// 		scene = document.Scenes[0]

// 		if len(scene.Nodes) != 1 {
// 			panic("single entity asset loading only supports a singular root entity")
// 		}
// 		node := scene.Nodes[0]

// 		entity := InstantiateEntity(document.Name)
// 		entity.MeshComponent = &MeshComponent{MeshHandle: handle, Transform: mgl64.Ident4(), Visible: true, ShadowCasting: true}
// 		var vertices []modelspec.Vertex
// 		VerticesFromNode(node, document, &vertices)
// 		entity.InternalBoundingBox = collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
// 		SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
// 		SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
// 		SetScale(entity, utils.Vec3F32ToF64(node.Scale))

// 		if len(document.Animations) > 0 {
// 			entity.Animation = NewAnimationComponent(document.Name, ml)
// 		}
// 		spawnedEntities = append(spawnedEntities, entity)
// 	} else {
// 		parent := InstantiateEntity(fmt.Sprintf("%s-parent", document.Name))
// 		spawnedEntities = append(spawnedEntities, parent)

// 		for _, scene := range document.Scenes {
// 			for _, node := range scene.Nodes {
// 				spawnedEntities = append(spawnedEntities, parseEntities(node, nil, document.Name, document, ml)...)
// 			}
// 		}

// 		var rootEntities []*Entity
// 		for _, e := range spawnedEntities {
// 			if e.Parent == nil {
// 				rootEntities = append(rootEntities, e)
// 			}
// 		}

// 		// only parent root entities
// 		for _, e := range rootEntities {
// 			if e.ID == parent.ID {
// 				continue
// 			}

// 			parent.Children[e.ID] = e
// 			e.Parent = parent
// 		}
// 	}

// 	for _, entity := range spawnedEntities {
// 		entity.Static = entityAsset.Static
// 		if entityAsset.Physics == nil {
// 			entity.Physics = &PhysicsComponent{}
// 		}
// 		if entityAsset.Collider != nil && entityAsset.Collider.TriMeshCollider {
// 			if entity.MeshComponent == nil {
// 				continue
// 			}
// 			meshHandle := entity.MeshComponent.MeshHandle
// 			primitives := ml.GetPrimitives(meshHandle)
// 			if entity.Collider == nil {
// 				if _, ok := ColliderGroupMap[ColliderGroup(entityAsset.Collider.ColliderGroup)]; !ok {
// 					panic(fmt.Sprintf("unrecognized collider group %s for document %s", entityAsset.Collider.ColliderGroup, document.Name))
// 				}
// 				entity.Collider = &ColliderComponent{ColliderGroup: ColliderGroupMap[ColliderGroup(entityAsset.Collider.ColliderGroup)]}
// 			}
// 			entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(MLPrimitivesTospecPrimitive(primitives))
// 		}
// 	}

// 	if len(spawnedEntities) > 0 {
// 		rootEntity := spawnedEntities[0]
// 		if entityAsset.Translation != nil {
// 			SetLocalPosition(rootEntity, *entityAsset.Translation)
// 		}
// 		if entityAsset.Rotation != nil {
// 			SetLocalRotation(rootEntity, *entityAsset.Rotation)
// 		}
// 		if entityAsset.Scale != nil {
// 			SetScale(rootEntity, *entityAsset.Scale)
// 		}
// 	}

// 	return spawnedEntities
// }

func (g *Client) InstantiateEntity(documentAsset assets.DocumentAsset) *entities.Entity {
	if !documentAsset.Config.SingleEntity {
		spawnedEntities := entities.CreateEntitiesFromDocument2(documentAsset, g.assetManager)
		for _, entity := range spawnedEntities {
			g.world.AddEntity(entity)
		}
		return spawnedEntities[0]
	}

	entityHandle := documentAsset.Config.Name
	document := g.AssetManager().GetDocument(entityHandle)
	handle := assets.NewSingleMeshHandle(entityHandle)
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

	primitives := g.AssetManager().GetPrimitives(handle)
	if len(primitives) > 0 {
		entity.Collider = &entities.ColliderComponent{
			ColliderGroup:   types.ColliderGroupMap[types.ColliderGroupTerrain],
			TriMeshCollider: collider.CreateTriMeshFromPrimitives(entities.MLPrimitivesTospecPrimitive(primitives)),
		}
	}

	g.world.AddEntity(entity)

	return entity
}

func (g *Client) BuildNavMesh(app renderiface.App, iterationCount int, walkableHeight int, climbableHeight int, minRegionArea int, sampleDist float64, maxError float64) {
	start := time.Now()
	defer func() {
		fmt.Println("BuildNavMesh completed in", time.Since(start))
	}()

	cs := app.RuntimeConfig().NavigationMeshCellSize
	ch := app.RuntimeConfig().NavigationMeshCellHeight
	walkableRadius := app.RuntimeConfig().NavigationMeshAgentRadius

	// walkableRadius /= cs
	// walkableHeight = (int(math.Ceil(float64(walkableHeight) / float64(ch))))
	// climbableHeight = (int(math.Floor(float64(climbableHeight) / float64(ch))))

	minVertex := mgl64.Vec3{-200.0, -5.0, -200.0}
	maxVertex := mgl64.Vec3{200.0, 80.0, 200.0}

	hfWidth := int((maxVertex.X()-minVertex.X())/float64(cs) + 0.5)
	hfHeight := int((maxVertex.Z()-minVertex.Z())/float64(cs) + 0.5)
	nmbb := collider.BoundingBox{MinVertex: minVertex, MaxVertex: maxVertex}

	hf := navmesh.NewHeightField(hfWidth, hfHeight, minVertex, maxVertex, float64(cs), float64(ch))
	var debugLines [][2]mgl64.Vec3

	world := app.World()
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

		primitives := app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
		transform := utils.Mat4F64ToF32(entities.WorldTransform(entity))
		up := mgl64.Vec3{0, 1, 0}

		for _, p := range primitives {
			for i := 0; i < len(p.Primitive.Vertices); i += 3 {
				v1 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i].Position.Vec4(1)).Vec3())
				v2 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+1].Position.Vec4(1)).Vec3())
				v3 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+2].Position.Vec4(1)).Vec3())

				tv1 := v2.Sub(v1)
				tv2 := v3.Sub(v2)

				normal := tv1.Cross(tv2).Normalize()
				isUp := normal.Dot(up) >= 0.7

				navmesh.RasterizeTriangle(v1, v2, v3, float64(cs), float64(ch), hf, isUp, climbableHeight)
			}
		}
	}

	if app.RuntimeConfig().NavigationMeshFilterLedgeSpans {
		navmesh.FilterLedgeSpans(walkableHeight, climbableHeight, hf)
	}

	if app.RuntimeConfig().NavigationMeshFilterLowHeightSpans {
		navmesh.FilterLowHeightSpans(walkableHeight, hf)
	}

	chf := navmesh.NewCompactHeightField(walkableHeight, climbableHeight, hf)

	navmesh.ErodeWalkableArea(chf, walkableRadius)
	navmesh.BuildDistanceField(chf)

	navmesh.BuildRegions(chf, iterationCount, minRegionArea, 1)
	contourSet := navmesh.BuildContours(chf, maxError, int(app.RuntimeConfig().NavigationmeshMaxEdgeLength))
	mesh := navmesh.BuildPolyMesh(contourSet)
	detailedMesh := navmesh.BuildDetailedPolyMesh(mesh, chf, app.RuntimeConfig())

	nm := &navmesh.NavigationMesh{
		HeightField:          hf,
		CompactHeightField:   chf,
		Volume:               nmbb,
		BlurredDistances:     chf.Distances,
		DebugLines:           debugLines,
		Invalidated:          true,
		InvalidatedTimestamp: int(time.Now().Unix()),
		ContourSet:           contourSet,
		Mesh:                 mesh,
		DetailedMesh:         detailedMesh,
	}

	g.navMesh = nm
}

func (g *Client) NavMesh() *navmesh.NavigationMesh {
	return g.navMesh
}

func (g *Client) CreateMaterial(material material.Material) {
	g.MaterialBrowser().AddMaterial(material)
}

func (g *Client) MaterialBrowser() *materialbrowser.MaterialBrowser {
	return g.project.MaterialBrowser
}

func (g *Client) FindPath(start, goal mgl64.Vec3) {
	g.navMesh.Invalidated = true
	c := navmesh.CompileNavMesh(g.navMesh)
	path := navmesh.FindPath(c, start, goal)

	navmesh.PATHPOLYGONS = make(map[int]bool)
	for _, p := range path {
		navmesh.PATHPOLYGONS[p] = true
	}
}

func (g *Client) CreateBatch() {
	g.renderSystem.CreateBatch()
}
