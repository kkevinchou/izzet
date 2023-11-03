package renderiface

import (
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

type App interface {
	AssetManager() *assets.AssetManager
	ModelLibrary() *modellibrary.ModelLibrary
	GetEditorCamera() *camera.Camera
	Prefabs() []*prefabs.Prefab
	NavMesh() *navmesh.NavigationMesh
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

	Serializer() *serialization.Serializer
	LoadWorld(string) bool
	SaveWorld(string)

	CollisionObserver() *observers.CollisionObserver
	Settings() *app.Settings
	Connect() error
	ConnectAndInitialize() error
	IsConnected() bool
	MetricsRegistry() *metrics.MetricsRegistry
	GetPlayerCamera() *entities.Entity

	StartAsyncServer()
	DisconnectAsyncServer()
	AsyncServerStarted() bool
	DisconnectClient()
}
