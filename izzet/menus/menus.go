package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/serialization"
)

type World interface {
	Serializer() *serialization.Serializer
	LoadSerializedWorld(world serialization.SerializedWorld)
}

func SetupMenuBar(world World) imgui.Vec2 {
	imgui.BeginMainMenuBar()
	size := imgui.WindowSize()
	if imgui.BeginMenu("File") {
		if imgui.MenuItem("Save") {
			world.Serializer().WriteOut("./scene.txt")
		}
		if imgui.MenuItem("Load") {
			serializedWorld := world.Serializer().ReadIn("./scene.txt")
			world.LoadSerializedWorld(serializedWorld)
			// entities := world.Serializer().DeserializeEntities(serializedWorld.Entities)

		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
