package menus

import (
	"fmt"
	"time"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/kitolib/utils"
)

var NM *navmesh.NavigationMesh

var map3D [][][]navmesh.Voxel2

func build(app renderiface.App, world renderiface.GameWorld) {
	// runtimeConfig := app.RuntimeConfig()
	// imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Build") {
		if imgui.MenuItemBool("Build Navigation Mesh") {
			fmt.Println("Build Navigation Mesh ")
			minVertex := mgl64.Vec3{-500, -250, -500}
			maxVertex := mgl64.Vec3{500, 250, 500}
			NM = navmesh.New(app, world, minVertex, maxVertex)

			// triangles := []navmesh.Triangle2{{Vertices: [3]mgl64.Vec3{
			// 	{0, 0, 0},
			// 	{100, 5, -10},
			// 	{60, 100, -50},
			// }}}

			// NM.Voxelize2(triangles)

			vxs := int(maxVertex.X() - minVertex.X())
			vys := int(maxVertex.Y() - minVertex.Y())
			vzs := int(maxVertex.Z() - minVertex.Z())

			if len(map3D) > 0 {
				// for x := range vxs {
				// 	for y := range vys {
				// 		for z := range vzs {
				// 			map3D[x][y][z].Filled = false
				// 		}
				// 	}
				// }
			} else {
				map3D = make([][][]navmesh.Voxel2, vxs)
				for x := range vxs {
					map3D[x] = make([][]navmesh.Voxel2, vys)
					for y := range vys {
						map3D[x][y] = make([]navmesh.Voxel2, vzs)
					}
				}
			}

			count := 0
			for _, entity := range world.Entities() {
				if entity.MeshComponent == nil {
					continue
				}

				if entity.GetID() != 541 {
					continue
				}

				primitives := app.ModelLibrary().GetPrimitives(entity.MeshComponent.MeshHandle)
				transform := utils.Mat4F64ToF32(entities.WorldTransform(entity))

				start := time.Now()
				for _, p := range primitives {
					for i := 0; i < len(p.Primitive.Vertices); i += 3 {
						v1 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i].Position.Vec4(1)).Vec3())
						v2 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+1].Position.Vec4(1)).Vec3())
						v3 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+2].Position.Vec4(1)).Vec3())

						c := navmesh.RasterizeTriangle(
							int(v1.X()),
							int(v1.Y()),
							int(v1.Z()),
							int(v2.X()),
							int(v2.Y()),
							int(v2.Z()),
							int(v3.X()),
							int(v3.Y()),
							int(v3.Z()),
							map3D,
						)

						count += c
					}
				}
				fmt.Println(fmt.Sprintf("rasterized %d voxels in", count), time.Since(start).Seconds())
			}

			var debugVoxels []mgl64.Vec3

			for i := range vxs {
				for j := range vys {
					for k := range vzs {
						_ = i
						_ = j
						_ = k
						voxel := map3D[i][j][k]
						if map3D[i][j][k].Filled {
							if voxel.Filled {
								debugVoxels = append(debugVoxels, mgl64.Vec3{float64(voxel.X), float64(voxel.Y), float64(voxel.Z)})
								// debugVoxels = append(debugVoxels, mgl64.Vec3{float64(map3D[i][j][k].X), float64(map3D[i][j][k].Y), float64(map3D[i][j][k].Z)})
							}
						}
					}
				}
			}
			NM.DebugVoxels = debugVoxels
		}

		imgui.EndMenu()
	}
}
