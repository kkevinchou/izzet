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
			// save project before attempting to connect. There may be imported assets
			// that have not been persisted to our project's disk yet which will cause them
			// to fail to load on connect.
			//
			// i might revisit this in the future so that importing always immediately saves
			// the asset to disk
			app.SaveProject()
			err := app.Connect()
			if err != nil {
				fmt.Println(err)
			}
		}

		if imgui.MenuItemBoolV("Disconnect Client", "", false, app.IsConnected()) {
			app.DisconnectClient()
		}

		if imgui.MenuItemBoolV("Start Async Server", "", app.AsyncServerStarted(), !app.AsyncServerStarted()) {
			// save project before attempting to start the server. There may be imported assets
			// that have not been persisted to our project's disk yet which will cause them
			// to fail to load on connect.
			//
			// i might revisit this in the future so that importing always immediately saves
			// the asset to disk
			app.SaveProject()
			app.StartAsyncServer()
		}

		if imgui.MenuItemBoolV("Stop Async Server", "", false, app.AsyncServerStarted()) {
			app.DisconnectAsyncServer()
		}

		imgui.EndMenu()
	}
}
