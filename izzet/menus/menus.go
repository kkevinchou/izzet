package menus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/navmesh"
)

type World interface {
	SaveWorld(string)
	LoadWorld(string)
	SetShowImguiDemo(bool)
	ShowImguiDemo() bool
	NavMesh() *navmesh.NavigationMesh
	AddEntity(entity *entities.Entity)
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
		imgui.Checkbox("ShowImguiDemoCheckbox", &val)
		world.SetShowImguiDemo(val)

		if imgui.MenuItem("Bake Navigation Mesh") {
			world.NavMesh().BakeNavMesh()
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
