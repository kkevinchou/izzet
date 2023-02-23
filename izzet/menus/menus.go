package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/panels"
)

type World interface {
	SaveWorld()
	LoadWorld()
	AddEntity(entity *entities.Entity)
}

func SetupMenuBar(world World) imgui.Vec2 {
	imgui.BeginMainMenuBar()
	size := imgui.WindowSize()
	if imgui.BeginMenu("File") {
		if imgui.MenuItem("Save") {
			world.SaveWorld()
		}
		if imgui.MenuItem("Load") {
			world.LoadWorld()
		}
		if imgui.MenuItem("Show Debug") {
			panels.ShowDebug = !panels.ShowDebug
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
