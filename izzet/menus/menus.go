package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
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
			// entities := world.Serializer().DeserializeEntities(serializedWorld.Entities)
		}
		if imgui.MenuItem("Add Collision Volume") {
			world.AddEntity(entities.CreateCube())
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
