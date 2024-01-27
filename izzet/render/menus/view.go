package menus

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func view(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 200})
	if imgui.BeginMenu("View") {
		if imgui.MenuItemBoolV("Show Colliders", "", runtimeConfig.RenderColliders, true) {
			runtimeConfig.RenderColliders = !runtimeConfig.RenderColliders
		}

		if imgui.MenuItemBoolV("Show UI", "", runtimeConfig.UIEnabled, true) {
			app.ConfigureUI(!runtimeConfig.UIEnabled)
		}

		if imgui.MenuItemBoolV("ShowImguiDemo", "", app.ShowImguiDemo(), true) {
			app.SetShowImguiDemo(!app.ShowImguiDemo())
		}

		imgui.EndMenu()
	}
}
