package panels

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/types"
)

var open bool

func BuildTabsSet(app renderiface.App, renderContext RenderContext) {
	imgui.BeginChildStrV("Right Window", imgui.Vec2{}, imgui.ChildFlagsNone, imgui.WindowFlagsNoBringToFrontOnFocus)
	if imgui.BeginTabBar("Main") {
		if imgui.BeginTabItem("Details") {
			entityProps(app.SelectedEntity(), app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Scene Graph") {
			sceneGraph(app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("World") {
			worldProps(app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Stats") {
			stats(app, renderContext)
			imgui.EndTabItem()
		}
		// if imgui.BeginTabItem("HUD") {
		// 	hud(app, renderContext)
		// 	imgui.EndTabItem()
		// }
		if app.RuntimeConfig().WindowEnablePostProcessing {
			if imgui.BeginTabItem("Rendering") {
				rendering(app)
				imgui.EndTabItem()
			}
		}
		if app.AppMode() == types.AppModePlay {
			if imgui.BeginTabItem("Controls") {
				controls(app, renderContext)
				imgui.EndTabItem()
			}
		}
		imgui.EndTabBar()
	}

	imgui.EndChild()
}
