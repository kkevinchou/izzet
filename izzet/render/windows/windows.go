package windows

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var (
	tableFlags imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

func RenderWindows(app renderiface.App) {
	renderMaterialWindow(app)
}
