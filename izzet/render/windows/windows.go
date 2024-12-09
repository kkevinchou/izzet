package windows

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var (
	tableFlags imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

func RenderWindows(app renderiface.App) {
	renderMaterialWindow(app)
	renderAnimationEditorWindow(app)
}
