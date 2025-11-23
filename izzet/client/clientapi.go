package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/assets/loaders"
	"github.com/kkevinchou/izzet/izzet/client/edithistory"
	"github.com/kkevinchou/izzet/izzet/client/editorcamera"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/system/clientsystems"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/world"
)

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

	world, err := serialization.ReadFromFile(filepath, g.assetManager)
	if err != nil {
		fmt.Println("failed to load world", filepath, err)
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
	entity.SetNextID(maxID + 1)

	g.SelectEntity(nil)
	g.world = world
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

func (g *Client) StartLiveWorld() {
	g.IsClient()
	if g.AppMode() != types.AppModeEditor {
		return
	}
	g.appMode = types.AppModePlay
	g.editorWorld = g.world

	var buffer bytes.Buffer
	err := serialization.Write(g.world, &buffer)
	if err != nil {
		panic(err)
	}

	liveWorld, err := serialization.Read(&buffer, g.assetManager)
	if err != nil {
		panic(err)
	}

	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	g.SelectEntity(nil)
	g.world = liveWorld
}

func (g *Client) StopLiveWorld() {
	if g.AppMode() != types.AppModePlay {
		return
	}
	g.appMode = types.AppModeEditor
	// TODO: more global state that needs to be cleaned up still, mostly around entities that are selected
	g.SelectEntity(nil)
	g.world = g.editorWorld
}

func (g *Client) AppMode() types.AppMode {
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

	fmt.Println("connected to server hosting project", message.ProjectName)

	g.initializeAssetManagerWithProject(message.ProjectName)
	g.playerID = message.PlayerID
	g.connection = conn
	g.networkMessages = make(chan network.MessageTransport, 100)

	// initialize the player's camera and playerEntity

	playerEntity, err := serialization.DeserializeEntity(message.EntityBytes, g.AssetManager())
	if err != nil {
		fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
	}
	g.world.AddEntity(playerEntity)

	camera, err := serialization.DeserializeEntity(message.CameraBytes, g.AssetManager())
	if err != nil {
		fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
	}
	g.world.AddEntity(camera)

	g.SetPlayerCamera(camera)
	g.SetPlayerEntity(playerEntity)

	fmt.Println("CLIENT player id", playerEntity.GetID(), "camera id", camera.GetID())

	world, err := serialization.Read(bytes.NewReader(message.Snapshot), g.assetManager)
	if err != nil {
		return err
	}

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
func (g *Client) SetPlayerEntity(entity *entity.Entity) {
	g.playerEntity = entity
}
func (g *Client) SetPlayerCamera(entity *entity.Entity) {
	g.playerCamera = entity
}
func (g *Client) GetPlayerEntity() *entity.Entity {
	return g.playerEntity
}
func (g *Client) GetPlayerCamera() *entity.Entity {
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

		world, err := serialization.Read(&worldBytes, g.assetManager)
		if err != nil {
			panic(err)
		}

		serverApp := server.NewWithWorld(world, g.project.Name)
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
	newConfig := g.CopyDocumentToProjectFolder(config)
	g.assetManager.LoadAndRegisterDocument(newConfig)
}

func (g *Client) CopyDocumentToProjectFolder(config assets.AssetConfig) assets.AssetConfig {
	peripheralFiles, err := loaders.GetPeripheralFiles(config.FilePath)
	if err != nil {
		panic(err)
	}

	sourceFilePaths := []string{config.FilePath}
	for _, peripheralFilePath := range peripheralFiles {
		sourceFilePaths = append(sourceFilePaths, filepath.Join(filepath.Dir(config.FilePath), peripheralFilePath))
	}

	contentDir := filepath.Join(settings.ProjectsDirectory, g.project.Name, "content")
	newConfig := config
	newConfig.FilePath = filepath.Join(contentDir, filepath.Base(config.FilePath))

	sourceRootDir := filepath.Dir(config.FilePath)
	err = copySourceFiles(sourceFilePaths, sourceRootDir, contentDir)
	if err != nil {
		panic(err)
	}
	return newConfig
}

func (g *Client) LoadDefaultAssets() {
	// default materials

	defaultMaterial := modelspec.MaterialSpecification{
		PBRMaterial: modelspec.PBRMaterial{
			PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
				BaseColorTextureName: settings.DefaultTexture,
				BaseColorFactor:      mgl32.Vec4{1, 1, 1, 1},
				RoughnessFactor:      .55,
				MetalicFactor:        0,
			},
		},
	}

	g.assetManager.CreateMaterialWithHandle("default material", defaultMaterial, assets.DefaultMaterialHandle)

	whiteMaterial := modelspec.MaterialSpecification{
		PBRMaterial: modelspec.PBRMaterial{
			PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
				BaseColorFactor: mgl32.Vec4{1, 1, 1, 1},
				RoughnessFactor: .55,
				MetalicFactor:   0,
			},
		},
	}
	g.assetManager.CreateMaterialWithHandle("white material", whiteMaterial, assets.WhiteMaterialHandle)

	// default models

	var subDirectories []string = []string{"gltf"}
	extensions := map[string]any{
		".gltf": nil,
	}
	fileMetaData := utils.GetFileMetaData(settings.BuiltinAssetsDir, subDirectories, extensions)
	for _, metaData := range fileMetaData {
		if strings.HasPrefix(metaData.Name, "_") {
			continue
		}

		config := assets.AssetConfig{
			Name:          metaData.Name,
			FilePath:      metaData.Path,
			ColliderType:  string(types.ColliderTypeMesh),
			ColliderGroup: string(types.ColliderGroupPlayer),
			SingleEntity:  true,
		}
		g.ImportAsset(config)
	}
}

func (g *Client) Shutdown() {
	g.gameOver = true
}

func (g *Client) ConfigureUI(enabled bool) {
	g.runtimeConfig.UIEnabled = enabled
	g.renderSystem.ReinitializeFrameBuffers()
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

func (g *Client) SelectEntity(entity *entity.Entity) {
	g.selectedEntity = entity
}

func (g *Client) SelectedEntity() *entity.Entity {
	return g.selectedEntity
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
	for _, e := range world.Entities() {
		if e.MeshComponent == nil {
			continue
		}
		if !e.HasBoundingBox() {
			continue
		}

		ebb := e.BoundingBox()

		if ebb.MaxVertex.X() < nmbb.MinVertex.X() || ebb.MinVertex.X() > nmbb.MaxVertex.X() {
			continue
		}
		if ebb.MaxVertex.Y() < nmbb.MinVertex.Y() || ebb.MinVertex.Y() > nmbb.MaxVertex.Y() {
			continue
		}
		if ebb.MaxVertex.Z() < nmbb.MinVertex.Z() || ebb.MinVertex.Z() > nmbb.MaxVertex.Z() {
			continue
		}

		primitives := app.AssetManager().GetPrimitives(e.MeshComponent.MeshHandle)
		transform := utils.Mat4F64ToF32(entity.WorldTransform(e))
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

func (g *Client) FindPath(start, goal mgl64.Vec3) {
	g.navMesh.Invalidated = true
	c := navmesh.CompileNavMesh(g.navMesh)
	path := navmesh.FindPath(c, start, goal)

	navmesh.PATHPOLYGONS = make(map[int]bool)
	for _, p := range path {
		navmesh.PATHPOLYGONS[p] = true
	}
}

func (g *Client) SetupBatchedStaticRendering() {
	g.renderSystem.SetupBatchedStaticRendering()
}

func (g *Client) PredictionDebugLogging() bool {
	return g.predictionDebugLogging
}

func (g *Client) SetPredictionDebugLogging(value bool) {
	g.predictionDebugLogging = value
}

func (g *Client) ResetApp() {
	g.world = world.New()
	g.editorWorld = world.New()
	g.initialize()
}

func (g *Client) NewProject(name string) {
	g.project = &Project{Name: name}
	g.InitializeProjectFolders(name)
	g.ResetApp()
	g.assetManager = assets.NewAssetManager(true)
	g.LoadDefaultAssets()
	g.SelectEntity(nil)

	// set up the default scene

	cube := entity.CreateCube(g.AssetManager(), 1)
	cube.Material = &entity.MaterialComponent{MaterialHandle: assets.DefaultMaterialHandle}
	entity.SetLocalPosition(cube, mgl64.Vec3{0, -1, 0})
	entity.SetScale(cube, mgl64.Vec3{7, 2, 7})
	g.World().AddEntity(cube)

	directionalLight := entity.CreateDirectionalLight()
	directionalLight.LightInfo.Diffuse3F = [3]float32{1, 1, 1}
	directionalLight.LightInfo.Direction3F = [3]float32{-0.5, -1, 1}
	directionalLight.Name = "directional_light"
	directionalLight.LightInfo.PreScaledIntensity = 4
	entity.SetLocalPosition(directionalLight, mgl64.Vec3{0, 20, 0})
	g.World().AddEntity(directionalLight)

	g.SaveProjectAs(name)
}

func (g *Client) QueueCreateMaterialTexture(handle types.MaterialHandle) {
	g.renderSystem.QueueCreateMaterialTexture(handle)
}
