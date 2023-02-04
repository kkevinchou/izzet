package menus

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/serialization"
)

type World interface {
	Serializer() *serialization.Serializer
}

func SetupMenuBar(world World) imgui.Vec2 {
	imgui.BeginMainMenuBar()
	size := imgui.WindowSize()
	if imgui.BeginMenu("File") {
		if imgui.MenuItem("Open") {
		}
		if imgui.MenuItem("Save") {
			world.Serializer().WriteOut("")
		}
		imgui.EndMenu()
	}
	imgui.EndMainMenuBar()
	return size
}
