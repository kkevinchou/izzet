package geometry_test

import (
	"fmt"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/izzet/izzet/apputils"
)

func TestZeroError(t *testing.T) {
	point1 := mgl32.Vec3{1, 1, 1}
	point2 := mgl32.Vec3{2, 1, 1}
	point3 := mgl32.Vec3{1.5, 1, -1}

	plane, ok := geometry.PlaneFromVerts([3]mgl32.Vec3{point1, point2, point3})
	if !ok {
		t.Fail()
	}
	v := mgl32.Vec4{5, 1, 0, 1}

	q := geometry.ComputeErrorQuadric(plane)

	errorValue := geometry.ComputeQEM(v, q)
	fmt.Println(errorValue)

	if errorValue != 0 {
		t.Fail()
	}
}

func TestErrorIsSquared(t *testing.T) {
	point1 := mgl32.Vec3{1, 1, 1}
	point2 := mgl32.Vec3{2, 1, 1}
	point3 := mgl32.Vec3{1.5, 1, -1}

	plane, ok := geometry.PlaneFromVerts([3]mgl32.Vec3{point1, point2, point3})
	if !ok {
		t.Fail()
	}
	v := mgl32.Vec4{5, 3, 0, 1}

	q := geometry.ComputeErrorQuadric(plane)

	errorValue := geometry.ComputeQEM(v, q)
	fmt.Println(errorValue)

	if errorValue != 4 {
		t.Fail()
	}
}

func TestPlaneWithXNormal(t *testing.T) {
	point1 := mgl32.Vec3{1, 1, 1}
	point2 := mgl32.Vec3{1, 1, -1}
	point3 := mgl32.Vec3{1, 2, -1}

	plane, ok := geometry.PlaneFromVerts([3]mgl32.Vec3{point1, point2, point3})
	if !ok {
		t.Fail()
	}
	v := mgl32.Vec4{5, 0, 0, 1}

	q := geometry.ComputeErrorQuadric(plane)

	errorValue := geometry.ComputeQEM(v, q)
	fmt.Println(errorValue)

	if errorValue != 16 {
		t.Fail()
	}
}

func TestFindMinimumVertex(t *testing.T) {
	// the three planes should intersect at {1 , 0, 0}
	point1 := mgl32.Vec3{1, 0, 0}
	point2 := mgl32.Vec3{2, 0, 0}
	point3 := mgl32.Vec3{1.5, 1, 0}

	plane1, ok := geometry.PlaneFromVerts([3]mgl32.Vec3{point1, point2, point3})
	if !ok {
		t.Fail()
	}

	point1 = mgl32.Vec3{1, 0, 0}
	point2 = mgl32.Vec3{1, 0, -1}
	point3 = mgl32.Vec3{1, 1, -1}

	plane2, ok := geometry.PlaneFromVerts([3]mgl32.Vec3{point1, point2, point3})
	if !ok {
		t.Fail()
	}

	point1 = mgl32.Vec3{1, 0, 0}
	point2 = mgl32.Vec3{2, 0, 0}
	point3 = mgl32.Vec3{1.5, 0, -1}

	plane3, ok := geometry.PlaneFromVerts([3]mgl32.Vec3{point1, point2, point3})
	if !ok {
		t.Fail()
	}

	q1 := geometry.ComputeErrorQuadric(plane1)
	q2 := geometry.ComputeErrorQuadric(plane2)
	q3 := geometry.ComputeErrorQuadric(plane3)

	totalQ := q1.Add(q2).Add(q3)

	totalQ[3] = 0
	totalQ[7] = 0
	totalQ[11] = 0
	totalQ[15] = 1

	if totalQ.Det() == 0 {
		panic("WAT")
	}

	vHat := totalQ.Inv().Mul4x1(mgl32.Vec4{0, 0, 0, 1})

	fmt.Println(vHat)

	if !vec4ApproxEqualThreshold(vHat, mgl32.Vec4{1, 0, 0, 1}, 0.1) {
		t.Fail()
	}
}

func vec4ApproxEqualThreshold(v1 mgl32.Vec4, v2 mgl32.Vec4, threshold float32) bool {
	return v1.ApproxFuncEqual(v2, func(a, b float32) bool {
		return apputils.F32Abs(a-b) < threshold
	})
}
