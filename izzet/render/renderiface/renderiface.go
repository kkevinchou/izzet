package renderiface

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/world"
)

type App interface {
	AssetManager() *assets.AssetManager
	GetEditorCameraPosition() mgl64.Vec3
	GetEditorCameraRotation() mgl64.Quat
	CommandFrame() int

	StopLiveWorld()
	AppMode() types.AppMode

	// for panels
	Platform() platforms.Platform

	LoadProject(name string) bool
	NewProject(name string)

	CollisionObserver() *collisionobserver.CollisionObserver
	RuntimeConfig() *runtimeconfig.RuntimeConfig
	Connect() error
	ConnectAndInitialize() error
	IsConnected() bool
	GetPlayerCamera() *entity.Entity

	StartAsyncServer()
	DisconnectAsyncServer()
	AsyncServerStarted() bool
	DisconnectClient()

	GetServerStats() serverstats.ServerStats
	SaveProject() error
	SaveProjectAs(name string) error

	GetPlayerEntity() *entity.Entity
	ConfigureUI(enabled bool)
	WindowSize() (int, int)
	Minimized() bool
	WindowFocused() bool
	ImportAsset(config assets.AssetConfig)
	SelectEntity(entity *entity.Entity)
	SelectedEntity() *entity.Entity
	CreateEntitiesFromDocumentAsset(documentAsset assets.DocumentAsset) *entity.Entity
	BuildNavMesh(App, int, int, int, int, float64, float64)
	NavMesh() *navmesh.NavigationMesh
	World() *world.GameWorld

	GetFrameInput() input.Input
	FindPath(start, goal mgl64.Vec3)
	SetupBatchedStaticRendering()
	ResetApp()

	QueueCreateMaterialTexture(handle types.MaterialHandle)
}

type RenderContext interface {
	Width() int
	Height() int
	AspectRatio() float64
}

type GameWorld interface {
	Entities() []*entity.Entity
	AddEntity(entity *entity.Entity)
	GetEntityByID(id int) *entity.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
