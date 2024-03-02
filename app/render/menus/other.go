package menus

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/app/render/renderiface"
)

func other(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Other") {
		if imgui.MenuItemBoolV("Lock Rendering To Update Rate", "", runtimeConfig.LockRenderingToCommandFrameRate, true) {
			runtimeConfig.LockRenderingToCommandFrameRate = !runtimeConfig.LockRenderingToCommandFrameRate
		}
		imgui.EndMenu()
	}
}
