package panels

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/mode"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var open bool

func BuildTabsSet(app renderiface.App, renderContext RenderContext, ps []*prefabs.Prefab) {
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
		if app.RuntimeConfig().WindowEnablePostProcessing {
			if imgui.BeginTabItem("Rendering") {
				rendering(app)
				imgui.EndTabItem()
			}
		}
		if app.AppMode() == mode.AppModePlay {
			if imgui.BeginTabItem("Controls") {
				controls(app, renderContext)
				imgui.EndTabItem()
			}
		}
		imgui.EndTabBar()
	}

	imgui.EndChild()
}
