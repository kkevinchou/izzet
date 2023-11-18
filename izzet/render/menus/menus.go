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

func SetupMenuBar(app renderiface.App) imgui.Vec2 {
	// imgui.SetNextWindowSize(imgui.Vec2{imgui.WindowSize().X, 400})
	imgui.BeginMainMenuBar()
	size := imgui.WindowSize()

	file(app)
	view(app)
	multiplayer(app)

	imgui.EndMainMenuBar()
	return size
}
