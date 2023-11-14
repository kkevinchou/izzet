package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func edit(app renderiface.App) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 200})
	if imgui.BeginMenu("Edit") {
		if imgui.MenuItemV("Snap To Grid", "", runtimeConfig.SnapToGrid, true) {
			runtimeConfig.SnapToGrid = !runtimeConfig.SnapToGrid
		}
		imgui.EndMenu()
	}
}
