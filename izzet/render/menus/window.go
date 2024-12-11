package menus

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func window(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Window") {
		if imgui.MenuItemBoolV("Rendering", "", runtimeConfig.WindowEnablePostProcessing, true) {
			runtimeConfig.WindowEnablePostProcessing = !runtimeConfig.WindowEnablePostProcessing
		}
		imgui.EndMenu()
	}
}
