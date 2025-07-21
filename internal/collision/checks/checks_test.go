package checks_test

import (
	"fmt"
	"testing"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision"
	"github.com/kkevinchou/izzet/internal/collision/checks"
	"github.com/kkevinchou/izzet/internal/collision/collider"
)

// Lines are perpendicular with separation along origin
func TestClosestPointsLineVSLine(t *testing.T) {
	points, distance := checks.ClosestPointsLineVSLine(
		collider.Line{
			P1: mgl64.Vec3{-1, 1, 0},
			P2: mgl64.Vec3{1, 1, 0},
		},
		collider.Line{
			P1: mgl64.Vec3{0, -1, -1},
			P2: mgl64.Vec3{0, -1, 1},
		},
	)

	expectedPoint0 := mgl64.Vec3{0, 1, 0}
	if points[0] != expectedPoint0 {
		t.Errorf("expected first point to be %v but got %v\n", expectedPoint0, points[0])
	}
	expectedPoint1 := mgl64.Vec3{0, -1, 0}
	if points[1] != expectedPoint1 {
		t.Errorf("expected second point to be %v but got %v\n", expectedPoint1, points[1])
	}
	var expectedDistance float64 = 2
	if distance != expectedDistance {
		t.Errorf("expected distance to be %f but got %f\n", expectedDistance, distance)
	}
}

// closest point is P1 one of the line segment and the center of the triangle
func TestClosestPointsLineVsTriangle(t *testing.T) {
	line := collider.Line{
		P1: mgl64.Vec3{0, 1, -0.5},
		P2: mgl64.Vec3{0, 2, -1},
	}
	trianglePoints := [3]mgl64.Vec3{
		{0, 0, 0},
		{1, 0, -1},
		{-1, 0, -1},
	}

	triangle := collider.NewTriangle(trianglePoints)
	points, distance := checks.ClosestPointsLineVSTriangle(line, triangle)

	expectedPoint0 := mgl64.Vec3{0, 1, -0.5}
	if points[0] != expectedPoint0 {
		t.Errorf("expected first point to be %v but got %v\n", expectedPoint0, points[0])
	}
	expectedPoint1 := mgl64.Vec3{0, 0, -0.5}
	if points[1] != expectedPoint1 {
		t.Errorf("expected second point to be %v but got %v\n", expectedPoint1, points[1])
	}
	var expectedDistance float64 = 1
	if distance != expectedDistance {
		t.Errorf("expected distance to be %f but got %f\n", expectedDistance, distance)
	}
}

// closest point is P2 one of the line segment and the center of the triangle
func TestClosestPointsLineVsTriangle2(t *testing.T) {
	line := collider.Line{
		P1: mgl64.Vec3{0, 2, -1},
		P2: mgl64.Vec3{0, 1, -0.5},
	}
	trianglePoints := [3]mgl64.Vec3{
		{0, 0, 0},
		{1, 0, -1},
		{-1, 0, -1},
	}

	triangle := collider.NewTriangle(trianglePoints)
	points, distance := checks.ClosestPointsLineVSTriangle(line, triangle)

	expectedPoint0 := mgl64.Vec3{0, 1, -0.5}
	if points[0] != expectedPoint0 {
		t.Errorf("expected first point to be %v but got %v\n", expectedPoint0, points[0])
	}
	expectedPoint1 := mgl64.Vec3{0, 0, -0.5}
	if points[1] != expectedPoint1 {
		t.Errorf("expected second point to be %v but got %v\n", expectedPoint1, points[1])
	}
	var expectedDistance float64 = 1
	if distance != expectedDistance {
		t.Errorf("expected distance to be %f but got %f\n", expectedDistance, distance)
	}
}

func TestTriangleEdgeClosestToLine(t *testing.T) {
	line := collider.Line{
		P1: mgl64.Vec3{0, -1, -2},
		P2: mgl64.Vec3{0, 1, -2},
	}
	trianglePoints := [3]mgl64.Vec3{
		{0, 0, 0},
		{1, 0, -1},
		{-1, 0, -1},
	}

	triangle := collider.NewTriangle(trianglePoints)
	points, distance := checks.ClosestPointsLineVSTriangle(line, triangle)

	expectedPoint0 := mgl64.Vec3{0, 0, -2}
	if points[0] != expectedPoint0 {
		t.Errorf("expected first point to be %v but got %v\n", expectedPoint0, points[0])
	}
	expectedPoint1 := mgl64.Vec3{0, 0, -1}
	if points[1] != expectedPoint1 {
		t.Errorf("expected second point to be %v but got %v\n", expectedPoint1, points[1])
	}
	var expectedDistance float64 = 1
	if distance != expectedDistance {
		t.Errorf("expected distance to be %f but got %f\n", expectedDistance, distance)
	}
}

func TestCheckCollisionCapsuleTriangle(t *testing.T) {
	capsule := collider.Capsule{
		Radius: 1,
		Top:    mgl64.Vec3{0, 10, -0.5},
		Bottom: mgl64.Vec3{0, 0.5, -0.5},
	}

	trianglePoints := [3]mgl64.Vec3{
		{0, 0, 0},
		{1, 0, -1},
		{-1, 0, -1},
	}

	triangle := collider.NewTriangle(trianglePoints)
	contact, _ := collision.CheckCollisionCapsuleTriangle(capsule, triangle)

	// expectedNormal := mgl64.Vec3{0, 1, 0}
	// if contact.Normal != expectedNormal {
	// 	t.Errorf("expected contact normal to be %v but got %v", expectedNormal, contact.Normal)
	// }

	if contact.SeparatingDistance != 0.5 {
		t.Errorf("expected separating distance to be %f but got %f", 0.5, contact.SeparatingDistance)
	}

	// expectedContactPoint := mgl64.Vec3{0, 0, -0.5}
	// if contact.Point != expectedContactPoint {
	// 	t.Errorf("expected contact point to be %v but got %v", expectedContactPoint, contact.Point)
	// }
}

// negative separating vector
func TestNegativeSeparatingVector(t *testing.T) {
	capsule := collider.Capsule{
		Radius: 3,
		Top:    mgl64.Vec3{228.1377, 47.30595, -293.3103},
		Bottom: mgl64.Vec3{228.1377, 0.30595, -293.3103},
	}

	trianglePoints := [3]mgl64.Vec3{
		{610.2427978, 1, -731.148681},
		{-538.179199, 1, -731.148681},
		{-538.179199, 1, 713.603515},
	}

	triangle := collider.NewTriangle(trianglePoints)
	contact, _ := collision.CheckCollisionCapsuleTriangle(capsule, triangle)
	fmt.Println(contact.SeparatingVector)
}

func TestPartWayCapsule(t *testing.T) {
	capsule := collider.Capsule{
		Radius: 5,
		Top:    mgl64.Vec3{0, 9, 0},
		Bottom: mgl64.Vec3{0, 1, 0},
	}

	trianglePoints := [3]mgl64.Vec3{
		{5, 3, 5},
		{0, 3, -5},
		{-5, 3, 5},
	}

	triangle := collider.NewTriangle(trianglePoints)
	contact, _ := collision.CheckCollisionCapsuleTriangle(capsule, triangle)

	fmt.Println(contact.SeparatingVector)
}

func TestDot(t *testing.T) {
	v1 := mgl64.Vec3{0, 1, 1}
	v2 := mgl64.Vec3{100, 0, 0}
	fmt.Println(v1.Dot(v2) / v2.Len())
}

func TestParallelInfiniteLines(t *testing.T) {
	p1 := mgl64.Vec3{0, 1, 0}
	q1 := mgl64.Vec3{1, 1, 0}
	p2 := mgl64.Vec3{0, 0, 0}
	q2 := mgl64.Vec3{0, 0, -1}

	a, b, nonParallel := checks.ClosestPointsInfiniteLines(p1, q1, p2, q2)
	if !nonParallel {
		t.Error("lines should not be parallel")
	}

	if a.Sub(b).Len() != 1 {
		t.Errorf("unexpected vector length of %f", a.Sub(b).Len())
	}
}

func TestParallelInfiniteLineVSLinePointingAway(t *testing.T) {
	p1 := mgl64.Vec3{0, 0, 0}
	q1 := mgl64.Vec3{1, 0, 0}
	p2 := mgl64.Vec3{0, 0, 99}
	q2 := mgl64.Vec3{0, 0, 100}

	_, b, nonParallel := checks.ClosestPointsInfiniteLineVSLine(p1, q1, p2, q2)
	if !nonParallel {
		t.Error("lines should not be parallel")
	}

	expectedPoint := mgl64.Vec3{0, 0, 99}
	if !b.ApproxEqual(expectedPoint) {
		t.Errorf("expected %v but instead got %v", expectedPoint, b)
	}
}

func TestParallelInfiniteLineVSLinePointingTowards(t *testing.T) {
	p1 := mgl64.Vec3{0, 0, 0}
	q1 := mgl64.Vec3{1, 0, 0}
	p2 := mgl64.Vec3{0, 0, 100}
	q2 := mgl64.Vec3{0, 0, 99}

	_, b, nonParallel := checks.ClosestPointsInfiniteLineVSLine(p1, q1, p2, q2)
	if !nonParallel {
		t.Error("lines should not be parallel")
	}

	expectedPoint := mgl64.Vec3{0, 0, 99}
	if !b.ApproxEqual(expectedPoint) {
		t.Errorf("expected %v but instead got %v", expectedPoint, b)
	}
}

func TestParallelInfiniteLineVSLinePointingOverlap(t *testing.T) {
	p1 := mgl64.Vec3{0, 0, 0}
	q1 := mgl64.Vec3{1, 0, 0}
	p2 := mgl64.Vec3{0, 0, 100}
	q2 := mgl64.Vec3{0, 0, -100}

	_, b, nonParallel := checks.ClosestPointsInfiniteLineVSLine(p1, q1, p2, q2)
	if !nonParallel {
		t.Error("lines should not be parallel")
	}

	expectedPoint := mgl64.Vec3{0, 0, 0}
	if !b.ApproxEqual(expectedPoint) {
		t.Errorf("expected %v but instead got %v", expectedPoint, b)
	}
}

func TestProjectPointOnPlane(t *testing.T) {
	plane := collider.Plane{Point: mgl64.Vec3{}, Normal: mgl64.Vec3{0, 1, 0}}
	point := mgl64.Vec3{100, 100, 0}
	projectedPoint := checks.ProjectPointOnPlane(point, plane)

	expectedProjectedPoint := mgl64.Vec3{100, 0, 0}
	if expectedProjectedPoint != projectedPoint {
		t.Errorf("expected %v but instead got %v", expectedProjectedPoint, projectedPoint)
	}

	point = mgl64.Vec3{100, -100, 0}
	projectedPoint = checks.ProjectPointOnPlane(point, plane)
	expectedProjectedPoint = mgl64.Vec3{100, 0, 0}
	if expectedProjectedPoint != projectedPoint {
		t.Errorf("expected %v but instead got %v", expectedProjectedPoint, projectedPoint)
	}
}
