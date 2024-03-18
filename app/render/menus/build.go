package menus

import (
	"fmt"
	"time"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/izzet/app/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/utils"
)

var NM *navmesh.NavigationMesh

func build(app renderiface.App, world renderiface.GameWorld) {
	// runtimeConfig := app.RuntimeConfig()
	// imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("Build") {
		if imgui.MenuItemBool("Build Navigation Mesh") {
			start := time.Now()
			nm := buildNavMesh(app, world)
			nm.Invalidated = true
			NM = nm
			fmt.Println("rasterized voxels in", time.Since(start).Seconds())
			fmt.Printf("rasterized %d spans\n", nm.HeightField.SpanCount())
		}

		imgui.EndMenu()
	}
}

func buildNavMesh(app renderiface.App, world renderiface.GameWorld) *navmesh.NavigationMesh {
	minVertex := mgl64.Vec3{-500, -250, -500}
	maxVertex := mgl64.Vec3{500, 250, 500}

	vxs := int(maxVertex.X() - minVertex.X())
	vzs := int(maxVertex.Z() - minVertex.Z())

	hf := navmesh.NewHeightField(vxs, vzs, minVertex, maxVertex)
	var debugLines [][2]mgl64.Vec3

	for _, entity := range world.Entities() {
		if entity.MeshComponent == nil {
			continue
		}

		primitives := app.ModelLibrary().GetPrimitives(entity.MeshComponent.MeshHandle)
		transform := utils.Mat4F64ToF32(entities.WorldTransform(entity))
		up := mgl64.Vec3{0, 1, 0}

		// rasterize triangles
		for _, p := range primitives {
			for i := 0; i < len(p.Primitive.Vertices); i += 3 {
				v1 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i].Position.Vec4(1)).Vec3())
				v2 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+1].Position.Vec4(1)).Vec3())
				v3 := utils.Vec3F32ToF64(transform.Mul4x1(p.Primitive.Vertices[i+2].Position.Vec4(1)).Vec3())

				debugLines = append(debugLines, [2]mgl64.Vec3{v1, v2})
				debugLines = append(debugLines, [2]mgl64.Vec3{v2, v3})
				debugLines = append(debugLines, [2]mgl64.Vec3{v3, v1})

				tv1 := v2.Sub(v1)
				tv2 := v3.Sub(v2)

				normal := tv1.Cross(tv2)
				if normal.LenSqr() > 0 {
					normal = normal.Normalize()
				}
				if normal.Dot(up) > 0.8 {
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
				}
			}
		}
	}

	// navmesh.FilterLowHeightSpans(500, hf)
	chf := navmesh.NewCompactHeightField(1, 1, hf)
	navmesh.BuildDistanceField(chf)
	navmesh.BuildRegions(chf)

	return &navmesh.NavigationMesh{
		HeightField:        hf,
		CompactHeightField: chf,
		Volume:             collider.BoundingBox{MinVertex: minVertex, MaxVertex: maxVertex},
		BlurredDistances:   chf.Distances(),
		DebugLines:         debugLines,
	}
}
