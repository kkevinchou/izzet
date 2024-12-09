package menus

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func multiplayer(app renderiface.App) {
	imgui.SetNextWindowSize(imgui.Vec2{X: 200})
	if imgui.BeginMenu("Multiplayer") {
		if imgui.MenuItemBoolV("Connect Client", "", app.IsConnected(), !app.IsConnected()) {
			err := app.ConnectAndInitialize()
			if err != nil {
				fmt.Println(err)
			}
		}

		if imgui.MenuItemBoolV("Disconnect Client", "", false, app.IsConnected()) {
			app.DisconnectClient()
		}

		if imgui.MenuItemBoolV("Start Async Server", "", app.AsyncServerStarted(), !app.AsyncServerStarted()) {
			app.StartAsyncServer()
		}

		if imgui.MenuItemBoolV("Stop Async Server", "", false, app.AsyncServerStarted()) {
			app.DisconnectAsyncServer()
		}

		imgui.EndMenu()
	}
}
