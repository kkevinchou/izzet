package renderiface

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/contentbrowser"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/observers"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/kitolib/assets"
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
	Platform() platforms.Platform

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
	SaveProject()
	SaveProjectAs(name string)

	GetPlayerEntity() *entities.Entity
	ConfigureUI(enabled bool)
	WindowSize() (int, int)
	Minimized() bool
	WindowFocused() bool
	ContentBrowser() *contentbrowser.ContentBrowser
	ImportToContentBrowser(assetFilePath string)
	SelectEntity(entity *entities.Entity)
	SelectedEntity() *entities.Entity
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
}
