package menus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/assets"
)

type World interface {
	SaveWorld(string)
	LoadWorld(string)
	SetShowImguiDemo(bool)
	ShowImguiDemo() bool
	NavMesh() *navmesh.NavigationMesh
	AddEntity(entity *entities.Entity)
	AssetManager() *assets.AssetManager
}

var worldName string = "scene"

var selectedWorldName string = ""

func SetupMenuBar(world World) imgui.Vec2 {
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
			world.SaveWorld(worldName)
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
			world.LoadWorld(selectedWorldName)
		}

		val := world.ShowImguiDemo()
		showImguiLabel := "ShowImguiDemo"
		if val {
			texture := panels.CreateUserSpaceTextureHandle(world.AssetManager().GetTexture("check-mark").ID)
			size := imgui.Vec2{X: 15, Y: 15}
			// invert the Y axis since opengl vs texture coordinate systems differ
			// https://learnopengl.com/Getting-started/Textures
			imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
			imgui.SameLine()
		}

		if imgui.MenuItem(showImguiLabel) {
			world.SetShowImguiDemo(!val)
		}

		if imgui.MenuItem("Bake Navigation Mesh") {
			world.NavMesh().BakeNavMesh()
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
