package menus

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/navmesh"
)

var NM *navmesh.NavigationMesh

func build(app renderiface.App, world renderiface.GameWorld) {
	// runtimeConfig := app.RuntimeConfig()
	// imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Build") {
		if imgui.MenuItemBool("Build Navigation Mesh") {
			fmt.Println("Build Navigation Mesh ")
			NM = navmesh.New(app, world)

			triangles := []navmesh.Triangle2{{Vertices: [3]mgl64.Vec3{
				{0, 0, 0},
				{100, 5, -10},
				{60, 100, -50},
			}}}

			NM.Voxelize2(triangles)
		}
		imgui.EndMenu()
	}
}
