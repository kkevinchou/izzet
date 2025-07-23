package renderiface

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/mode"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

type App interface {
	AssetManager() *assets.AssetManager
	GetEditorCameraPosition() mgl64.Vec3
	GetEditorCameraRotation() mgl64.Quat
	Prefabs() []*prefabs.Prefab
	ResetNavMeshVAO()
	CommandFrame() int

	StartLiveWorld()
	StopLiveWorld()
	AppMode() mode.AppMode

	// for panels
	GetPrefabByID(id int) *prefabs.Prefab
	Platform() platforms.Platform

	LoadProject(name string) bool

	CollisionObserver() *collisionobserver.CollisionObserver
	RuntimeConfig() *runtimeconfig.RuntimeConfig
	Connect() error
	ConnectAndInitialize() error
	IsConnected() bool
	MetricsRegistry() *metrics.MetricsRegistry
	GetPlayerCamera() *entities.Entity

	StartAsyncServer()
	DisconnectAsyncServer()
	AsyncServerStarted() bool
	DisconnectClient()

	GetServerStats() serverstats.ServerStats
	SaveProject(name string) error

	GetPlayerEntity() *entities.Entity
	ConfigureUI(enabled bool)
	WindowSize() (int, int)
	Minimized() bool
	WindowFocused() bool
	ImportAsset(config assets.AssetConfig)
	SelectEntity(entity *entities.Entity)
	SelectedEntity() *entities.Entity
	CreateEntitiesFromDocumentAsset(documentAsset assets.DocumentAsset) *entities.Entity
	BuildNavMesh(App, int, int, int, int, float64, float64)
	NavMesh() *navmesh.NavigationMesh
	World() *world.GameWorld

	GetFrameInput() input.Input
	FindPath(start, goal mgl64.Vec3)
	SetupBatchedStaticRendering()
	Reset()
}

type RenderContext interface {
	Width() int
	Height() int
	AspectRatio() float64
}

type GameWorld interface {
	Entities() []*entities.Entity
	AddEntity(entity *entities.Entity)
	GetEntityByID(id int) *entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
