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

			// triangles := []navmesh.Triangle2{{Vertices: [3]mgl64.Vec3{
			// 	{0, 0, 0},
			// 	{100, 5, -10},
			// 	{60, 100, -50},
			// }}}

			// NM.Voxelize2(triangles)

			vxs := 200
			vys := 200
			vzs := 200

			map3D := make([][][]float32, vxs)
			for x := range vxs {
				map3D[x] = make([][]float32, vys)
				for y := range vys {
					map3D[x][y] = make([]float32, vzs)
				}
			}

			// navmesh.Plinex(0, 0, 0, 50, 50, 0)

			v1 := mgl64.Vec3{0, 0, 45}
			v2 := mgl64.Vec3{45, 0, 45}
			v3 := mgl64.Vec3{45, 0, 0}
			navmesh.TriangleComp(int(v1.X()), int(v1.Y()), int(v1.Z()), int(v2.X()), int(v2.Y()), int(v2.Z()), int(v3.X()), int(v3.Y()), int(v3.Z()), map3D)
			NM.DebugLines = [][2]mgl64.Vec3{{v1, v2}, {v2, v3}, {v3, v1}}

			v1 = mgl64.Vec3{45, 0, 0}
			v2 = mgl64.Vec3{0, 0, 0}
			v3 = mgl64.Vec3{0, 0, 45}
			navmesh.TriangleComp(int(v1.X()), int(v1.Y()), int(v1.Z()), int(v2.X()), int(v2.Y()), int(v2.Z()), int(v3.X()), int(v3.Y()), int(v3.Z()), map3D)
			NM.DebugLines = append(NM.DebugLines, [][2]mgl64.Vec3{{v1, v2}, {v2, v3}, {v3, v1}}...)

			var debugVoxels []mgl64.Vec3

			for i := range vxs {
				for j := range vys {
					for k := range vzs {
						if map3D[i][j][k] == 1 {
							debugVoxels = append(debugVoxels, mgl64.Vec3{float64(i), float64(j), float64(k)})
						}
					}
				}
			}
			NM.DebugVoxels = debugVoxels
			// NM.DebugLines = [][2]mgl64.Vec3{{v1, {0, 100, 0}}}
		}
		imgui.EndMenu()
	}
}
