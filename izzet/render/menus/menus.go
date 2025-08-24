package menus

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var (
	tableFlags imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

var worldName string = ""
var selectedWorldName string = ""

func SetupMenuBar(app renderiface.App, renderContext RenderContext) {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 10, Y: 10})
	imgui.BeginMainMenuBar()

	file(app)
	view(app, renderContext)
	multiplayer(app)
	create(app)
	window(app)
	other(app)

	imgui.EndMainMenuBar()
	imgui.PopStyleVar()
}

type RenderContext interface {
	Width() int
	Height() int
	AspectRatio() float64
}
