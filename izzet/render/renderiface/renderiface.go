package renderiface

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/project"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

type App interface {
	AssetManager() *assets.AssetManager
	ModelLibrary() *modellibrary.ModelLibrary
	GetEditorCameraPosition() mgl64.Vec3
	GetEditorCameraRotation() mgl64.Quat
	Prefabs() []*prefabs.Prefab
	ResetNavMeshVAO()
	CommandFrame() int

	StartLiveWorld()
	StopLiveWorld()
	AppMode() app.AppMode

	// for panels
	GetPrefabByID(id int) *prefabs.Prefab
	Platform() *input.SDLPlatform

	SetShowImguiDemo(bool)
	ShowImguiDemo() bool

	LoadProject(name string) bool

	CollisionObserver() *observers.CollisionObserver
	RuntimeConfig() *app.RuntimeConfig
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
	GetProject() *project.Project
	SaveProject()
	SaveProjectAs(name string)

	GetPlayerEntity() *entities.Entity
}
