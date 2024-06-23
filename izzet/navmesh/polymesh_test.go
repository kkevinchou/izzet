package navmesh

import "testing"

func TestPolyMesh(t *testing.T) {
	// counter clockwise
	vertices := []SimplifiedVertex{
		{X: 100, Y: 0, Z: 0},
		{X: 100, Y: 0, Z: -100},
		{X: 50, Y: 0, Z: -150},
		{X: 0, Y: 0, Z: -100},
		{X: 0, Y: 0, Z: 0},
	}

	// // clockwise
	// vertices := []SimplifiedVertex{
	// 	{X: 0, Y: 0, Z: 0},
	// 	{X: 0, Y: 0, Z: -100},
	// 	{X: 50, Y: 0, Z: -150},
	// 	{X: 100, Y: 0, Z: -100},
	// 	{X: 100, Y: 0, Z: 0},
	// }

	tris := triangulate(vertices)

	if len(tris) != 3 {
		t.Errorf("expected 3 triangles but got %d", len(tris))
		return
	}
}

func TestLeft(t *testing.T) {
	a := SimplifiedVertex{X: 0, Y: 0, Z: 100}
	b := SimplifiedVertex{X: 0, Y: 0, Z: 0}
	c := SimplifiedVertex{X: -50, Y: 0, Z: 0}

	if !leftOn(a, b, c) {
		t.Error("leftOn should be true")
		return
	}

	if !left(a, b, c) {
		t.Error("left should be true")
		return
	}

	// c colinear
	c = SimplifiedVertex{X: 0, Y: 0, Z: 50}

	if !leftOn(a, b, c) {
		t.Error("leftOn should be true")
		return
	}

	if left(a, b, c) {
		t.Error("left should be false")
		return
	}
}

func TestInCone(t *testing.T) {
	a := SimplifiedVertex{X: 0, Y: 0, Z: 0}
	b := SimplifiedVertex{X: 0, Y: 0, Z: -100}
	c := SimplifiedVertex{X: -100, Y: 0, Z: -100}
	d := SimplifiedVertex{X: -100, Y: 0, Z: 0}

	vertices := []SimplifiedVertex{a, b, c, d}
	var indices []Index
	for i := range len(vertices) {
		indices = append(indices, Index{index: i})
	}
	if !inCone(0, 2, len(vertices), vertices, indices) {
		t.Error("inCone should be true")
		return
	}
}

func TestDiagonalie(t *testing.T) {

}
