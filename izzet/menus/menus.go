package menus

import (
	"github.com/go-gl/mathgl/mgl64"
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
		if imgui.MenuItem("Create Point Light") {
			lightInfo := &entities.LightInfo{
				Type:    1,
				Diffuse: mgl64.Vec4{1, 1, 1, 4000},
			}
			light := entities.CreateLight(lightInfo)
			world.AddEntity(light)
			panels.SelectEntity(light)
		}
		if imgui.MenuItem("Show Debug") {
			panels.ShowDebug = !panels.ShowDebug
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
