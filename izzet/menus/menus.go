package menus

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
)

type World interface {
	SaveWorld()
	LoadWorld()
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
			fmt.Println("yo")
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
