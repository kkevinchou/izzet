package menus

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
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
