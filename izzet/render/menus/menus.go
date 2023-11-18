package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
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

func SetupMenuBar(app renderiface.App) {
	imgui.BeginMainMenuBar()

	file(app)
	view(app)
	multiplayer(app)

	imgui.EndMainMenuBar()
}
