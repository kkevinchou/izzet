package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func view(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 200})
	if imgui.BeginMenu("View") {
		if imgui.MenuItemV("Show Colliders", "", runtimeConfig.RenderColliders, true) {
			runtimeConfig.RenderColliders = !runtimeConfig.RenderColliders
		}

		if imgui.MenuItemV("Show UI", "", runtimeConfig.UIEnabled, true) {
			runtimeConfig.UIEnabled = !runtimeConfig.UIEnabled
			app.ReinitializeFrameBuffers()
		}

		if imgui.MenuItemV("ShowImguiDemo", "", app.ShowImguiDemo(), true) {
			app.SetShowImguiDemo(!app.ShowImguiDemo())
		}

		imgui.EndMenu()
	}
}
