package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

// func TestRasterization(t *testing.T) {
// 	v0 := mgl64.Vec3{0, 0, 0}
// 	v1 := mgl64.Vec3{0, 0, 1}
// 	v2 := mgl64.Vec3{0, 1, 0}

// 	minVertex := mgl64.Vec3{-10, -10, -10}
// 	maxVertex := mgl64.Vec3{10, 10, 10}
// 	vxs := int(maxVertex.X() - minVertex.X())
// 	vzs := int(maxVertex.Z() - minVertex.Z())
// 	hf := NewHeightField(vxs, vzs, minVertex, maxVertex)

// 	RasterizeTriangle2(v0, v1, v2, 1, 1, 1, hf, true)
// 	fmt.Println("HI")
// }

func TestRasterization(t *testing.T) {
	v0 := mgl64.Vec3{0, 0, 0}
	v1 := mgl64.Vec3{0, 0, 5}
	v2 := mgl64.Vec3{0, 5, 0}

	minVertex := mgl64.Vec3{-10, -10, -10}
	maxVertex := mgl64.Vec3{10, 10, 10}
	vxs := int(maxVertex.X() - minVertex.X())
	vzs := int(maxVertex.Z() - minVertex.Z())
	hf := NewHeightField(vxs, vzs, minVertex, maxVertex)

	RasterizeTriangle2(v0, v1, v2, 1, 1, 1, hf, true, 10)
}
