package menus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/assets"
)

type App interface {
	SaveWorld(string)
	LoadWorld(string)
	SetShowImguiDemo(bool)
	ShowImguiDemo() bool
	NavMesh() *navmesh.NavigationMesh
	AssetManager() *assets.AssetManager

	StartLiveWorld()
	StopLiveWorld()
}

var worldName string = "scene"

var selectedWorldName string = ""

func SetupMenuBar(app App) imgui.Vec2 {
	imgui.BeginMainMenuBar()
	size := imgui.WindowSize()
	if imgui.BeginMenu("File") {
		imgui.PushID("World Name")
		imgui.InputText("", &worldName)
		imgui.PopID()

		imgui.SameLine()
		if imgui.Button("Save") {
			fmt.Println("Save to", worldName)
			// if imgui.MenuItem("Save") {
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
			app.LoadWorld(selectedWorldName)
		}

		val := app.ShowImguiDemo()
		showImguiLabel := "ShowImguiDemo"
		if val {
			texture := panels.CreateUserSpaceTextureHandle(app.AssetManager().GetTexture("check-mark").ID)
			size := imgui.Vec2{X: 20, Y: 20}

			// invert the Y axis since opengl vs texture coordinate systems differ
			// https://learnopengl.com/Getting-started/Textures
			imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
			imgui.SameLine()
		}

		if imgui.MenuItem(showImguiLabel) {
			app.SetShowImguiDemo(!val)
		}

		if imgui.MenuItem("Play Scene") {
			app.StartLiveWorld()
		}

		if imgui.MenuItem("Exit Scene") {
			app.StopLiveWorld()
		}

		if imgui.MenuItem("Bake Navigation Mesh") {
			app.NavMesh().BakeNavMesh()
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
