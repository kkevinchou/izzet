package menus

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func window(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Window") {
		if imgui.MenuItemBoolV("Show Post Processing Window", "", runtimeConfig.WindowEnablePostProcessing, true) {
			runtimeConfig.WindowEnablePostProcessing = !runtimeConfig.WindowEnablePostProcessing
		}
		imgui.EndMenu()
	}
}