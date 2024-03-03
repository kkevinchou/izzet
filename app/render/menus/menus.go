package menus

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var (
	ignoredJsonFiles map[string]any = map[string]any{
		"config.json":     true,
		"izzet_data.json": true,
	}
)

var worldName string = settings.DefaultProject
var selectedWorldName string = ""

func SetupMenuBar(app renderiface.App, world renderiface.GameWorld) {
	imgui.BeginMainMenuBar()

	file(app)
	view(app)
	build(app, world)
	multiplayer(app)
	other(app)

	imgui.EndMainMenuBar()
}
