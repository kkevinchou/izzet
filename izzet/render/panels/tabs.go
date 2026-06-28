package panels

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/appmode"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var open bool

func BuildTabsSet(app renderiface.App, renderContext RenderContext) {
	imgui.BeginChildStrV("Right Window", imgui.Vec2{}, imgui.ChildFlagsNone, imgui.WindowFlagsNoBringToFrontOnFocus)
	if imgui.BeginTabBar("Main") {
		if imgui.BeginTabItem("Details") {
			EntityProps(app.SelectedEntity(), app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Scene Graph") {
			SceneGraph(app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("World") {
			WorldProps(app)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Stats") {
			Stats(app, renderContext)
			imgui.EndTabItem()
		}
		if app.AppMode() == appmode.Play {
			if imgui.BeginTabItem("Controls") {
				controls(app, renderContext)
				imgui.EndTabItem()
			}
		}
		imgui.EndTabBar()
	}

	imgui.EndChild()
}
