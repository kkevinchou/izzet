package menus

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var (
	ignoredJsonFiles map[string]any = map[string]any{
		"config.json":     true,
		"izzet_data.json": true,
	}
	tableFlags imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

var worldName string = settings.DefaultProject
var selectedWorldName string = ""

func SetupMenuBar(app renderiface.App, renderContext RenderContext) {
	imgui.BeginMainMenuBar()

	file(app)
	view(app, renderContext)
	multiplayer(app)
	create(app)
	window(app)
	other(app)

	imgui.EndMainMenuBar()
}

type RenderContext interface {
	Width() int
	Height() int
	AspectRatio() float64
}
