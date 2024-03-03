package menus

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/navmesh"
)

var nm *navmesh.NavigationMesh

func build(app renderiface.App, world renderiface.GameWorld) {
	// runtimeConfig := app.RuntimeConfig()
	// imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Build") {
		if imgui.MenuItemBool("Build Navigation Mesh") {
			fmt.Println("Build Navigation Mesh ")
			nm = navmesh.New(app, world)
		}
		imgui.EndMenu()
	}
}
