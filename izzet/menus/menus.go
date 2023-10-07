package menus

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/panels"
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
		imgui.SameLine()
		if imgui.Button("Load") {
			fmt.Println("Load from", worldName)
			// if imgui.MenuItem("Load") {
			world.LoadWorld(worldName)
		}
		if imgui.MenuItem("Show Debug") {
			panels.ShowDebug = !panels.ShowDebug
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
