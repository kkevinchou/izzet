package menus

import (
	"fmt"
	"time"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/navmesh"
)

var NM *navmesh.NavigationMesh

var HeightField *navmesh.HeightField

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

			vxs := int(maxVertex.X() - minVertex.X())
			// vys := int(maxVertex.Y() - minVertex.Y())
			vzs := int(maxVertex.Z() - minVertex.Z())

			hf := navmesh.NewHeightField(vxs, vzs, minVertex, maxVertex)

			count := 0
			for _, entity := range world.Entities() {
				if entity.MeshComponent == nil {
					continue
				}

				if entity.GetID() != 541 {
					continue
				}

				v1 := mgl64.Vec3{0, 0, 0}
				v2 := mgl64.Vec3{200, 0, 0}
				v3 := mgl64.Vec3{100, 200, 0}
				start := time.Now()
				navmesh.RasterizeTriangle(
					int(v1.X()),
					int(v1.Y()),
					int(v1.Z()),
					int(v2.X()),
					int(v2.Y()),
					int(v2.Z()),
					int(v3.X()),
					int(v3.Y()),
					int(v3.Z()),
					hf,
				)

				// primitives := app.ModelLibrary().GetPrimitives(entity.MeshComponent.MeshHandle)
				// transform := utils.Mat4F64ToF32(entities.WorldTransform(entity))
				// for _, p := range primitives {
				// 	for i := 0; i < len(p.Primitive.Vertices); i += 3 {
				// 		v1 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i].Position.Vec4(1)).Vec3())
				// 		v2 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+1].Position.Vec4(1)).Vec3())
				// 		v3 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+2].Position.Vec4(1)).Vec3())

				// 		c := navmesh.RasterizeTriangle(
				// 			int(v1.X()),
				// 			int(v1.Y()),
				// 			int(v1.Z()),
				// 			int(v2.X()),
				// 			int(v2.Y()),
				// 			int(v2.Z()),
				// 			int(v3.X()),
				// 			int(v3.Y()),
				// 			int(v3.Z()),
				// 			hf,
				// 		)

				// 		count += c
				// 	}
				// }
				fmt.Println(fmt.Sprintf("rasterized %d voxels in", count), time.Since(start).Seconds())
			}
			HeightField = hf
		}

		imgui.EndMenu()
	}
}
