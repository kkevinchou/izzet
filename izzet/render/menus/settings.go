package menus

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/internal/platforms"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func settingsMenu(app renderiface.App) {
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Settings") {
		runtimeConfig := app.RuntimeConfig()

		windowMode, ok := app.Platform().(platforms.WindowModeController)
		fullscreen := false
		if ok {
			fullscreen = windowMode.Fullscreen()
		}

		if imgui.MenuItemBoolV("Fullscreen", "", fullscreen, ok) {
			if err := windowMode.SetFullscreen(!fullscreen); err != nil {
				fmt.Println("failed to update fullscreen mode:", err)
			}
		}

		if imgui.MenuItemBoolV("Lock Rendering To Update Rate", "", runtimeConfig.LockRenderingToCommandFrameRate, true) {
			runtimeConfig.LockRenderingToCommandFrameRate = !runtimeConfig.LockRenderingToCommandFrameRate
		}

		imgui.EndMenu()
	}
}
