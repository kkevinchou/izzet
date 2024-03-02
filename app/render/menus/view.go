package menus

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/app/render/renderiface"
)

func view(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("View") {
		if imgui.MenuItemBoolV("Show Colliders", "", runtimeConfig.ShowColliders, true) {
			runtimeConfig.ShowColliders = !runtimeConfig.ShowColliders
		}

		if imgui.MenuItemBoolV("Show Selection Bounding Box", "", runtimeConfig.ShowSelectionBoundingBox, true) {
			runtimeConfig.ShowSelectionBoundingBox = !runtimeConfig.ShowSelectionBoundingBox
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
