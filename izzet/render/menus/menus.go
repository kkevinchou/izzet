package menus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inkyblackness/imgui-go/v4"
	izzetapp "github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var (
	ignoredJsonFiles map[string]any = map[string]any{
		"config.json":     true,
		"izzet_data.json": true,
	}
)

var worldName string = "scene"

var selectedWorldName string = ""

func SetupMenuBar(app renderiface.App) imgui.Vec2 {
	settings := app.Settings()

	imgui.BeginMainMenuBar()
	size := imgui.WindowSize()
	if imgui.BeginMenu("File") {
		imgui.PushID("World Name")
		imgui.InputText("", &worldName)
		imgui.PopID()

		imgui.SameLine()
		if imgui.Button("Save") {
			fmt.Println("Save to", worldName)
			app.SaveWorld(worldName)
		}

		files, err := os.ReadDir(".")
		if err != nil {
			panic(err)
		}

		var savedWorlds []string
		for _, file := range files {
			extension := filepath.Ext(file.Name())
			if extension != ".json" {
				continue
			}

			if _, ok := ignoredJsonFiles[file.Name()]; ok {
				continue
			}

			name := file.Name()[0 : len(file.Name())-len(extension)]
			savedWorlds = append(savedWorlds, name)
		}

		if len(savedWorlds) == 0 {
			savedWorlds = append(savedWorlds, selectedWorldName)
		}

		if imgui.BeginCombo("", selectedWorldName) {
			for _, worldName := range savedWorlds {
				if imgui.Selectable(worldName) {
					selectedWorldName = worldName
				}
			}
			imgui.EndCombo()
		}
		imgui.SameLine()
		if imgui.Button("Load") {
			fmt.Println("Load from", selectedWorldName)
			if app.LoadWorld(selectedWorldName) {
				worldName = selectedWorldName
			}
		}

		// if imgui.MenuItem("Bake Navigation Mesh") {
		// 	app.NavMesh().BakeNavMesh()
		// }
		imgui.EndMenu()
	}

	imgui.SetNextWindowSize(imgui.Vec2{X: 200})
	if imgui.BeginMenu("Run") {
		if imgui.MenuItemV("Play Scene", "", app.AppMode() == izzetapp.AppModePlay, true) {
			app.StartLiveWorld()
		}

		if imgui.MenuItemV("Exit Scene", "", false, true) {
			app.StopLiveWorld()
		}

		imgui.EndMenu()
	}

	imgui.SetNextWindowSize(imgui.Vec2{X: 200})
	if imgui.BeginMenu("View") {
		if imgui.MenuItemV("Show Colliders", "", settings.RenderColliders, true) {
			settings.RenderColliders = !settings.RenderColliders
		}

		if imgui.MenuItemV("Show UI", "", settings.UIEnabled, true) {
			settings.UIEnabled = !settings.UIEnabled
		}

		if imgui.MenuItemV("ShowImguiDemo", "", app.ShowImguiDemo(), true) {
			app.SetShowImguiDemo(!app.ShowImguiDemo())
		}

		imgui.EndMenu()
	}

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

	imgui.EndMainMenuBar()
	return size
}
