package menus

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

func multiplayer(app renderiface.App) {
	imgui.SetNextWindowSize(imgui.Vec2{X: 200})
	if imgui.BeginMenu("Multiplayer") {
		if imgui.MenuItemV("Connect Client", "", app.IsConnected(), !app.IsConnected()) {
			err := app.ConnectAndInitialize()
			if err != nil {
				fmt.Println(err)
			}
		}

		if imgui.MenuItemV("Disconnect Client", "", false, app.IsConnected()) {
			app.DisconnectClient()
		}

		if imgui.MenuItemV("Start Async Server", "", app.AsyncServerStarted(), !app.AsyncServerStarted()) {
			app.StartAsyncServer()
		}

		if imgui.MenuItemV("Stop Async Server", "", false, app.AsyncServerStarted()) {
			app.DisconnectAsyncServer()
		}

		imgui.EndMenu()
	}
}
