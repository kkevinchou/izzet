package navmesh

import (
	"testing"
)

func TestBuildMeshAdjacency(t *testing.T) {
	polygons := []Polygon{
		{
			Verts:        []int{0, 1, 2},
			polyNeighbor: []int{-1, -1, -1},
		},
		{
			Verts:        []int{1, 0, 3},
			polyNeighbor: []int{-1, -1, -1},
		},
	}

	buildMeshAdjacency(polygons, 999999)

	// for the first polygon. edge (0, 1) should be a neighbor to polygon's edge (0, 1)
	if polygons[0].polyNeighbor[0] != 1 {
		t.Errorf("neighbor to edge at index 0 was %d instead of 1", polygons[0].polyNeighbor[0])
		return
	}
	for i := 1; i < len(polygons[0].polyNeighbor); i++ {
		neighbor := polygons[0].polyNeighbor[i]
		if neighbor != -1 {
			t.Errorf("neighbor to edge at index %d was %d instead of -1", i, neighbor)
			return
		}
	}

	if polygons[1].polyNeighbor[0] != 0 {
		t.Errorf("neighbor to edge at index 0 was %d instead of 0", polygons[1].polyNeighbor[0])
		return
	}
	for i := 1; i < len(polygons[1].polyNeighbor); i++ {
		neighbor := polygons[1].polyNeighbor[i]
		if neighbor != -1 {
			t.Errorf("neighbor to edge at index %d was %d instead of -1", i, neighbor)
			return
		}
	}
}

func TestBuildPolyMesh(t *testing.T) {
	contourSet := &ContourSet{
		Contours: []Contour{
			Contour{
				Verts: []SimplifiedVertex{
					{X: 100, Y: 0, Z: 0},
					{X: 100, Y: 0, Z: -100},
					{X: 50, Y: 0, Z: -150},
					{X: 0, Y: 0, Z: -100},
					{X: 0, Y: 0, Z: 0},
				},
			},
		},
	}

	mesh := BuildPolyMesh(contourSet)
	if mesh == nil {
		t.Fail()
	}

	if len(mesh.Vertices) != 5 {
		t.Errorf("expected 5 vertices but found %d", len(mesh.Vertices))
		return
	}
}

func TestBuildPolyMeshWithOverlappingVerts(t *testing.T) {
	contourSet := &ContourSet{
		Contours: []Contour{
			Contour{
				Verts: []SimplifiedVertex{
					{X: 100, Y: 0, Z: 0},
					{X: 100, Y: 0, Z: -100},
					{X: 50, Y: 0, Z: -150},
					{X: 0, Y: 0, Z: -100},
					{X: 0, Y: 0, Z: 0},
				},
			},
			Contour{
				Verts: []SimplifiedVertex{
					{X: 100, Y: 0, Z: 0},
					{X: 100, Y: 0, Z: -100},
					{X: 50, Y: 0, Z: -150},
					{X: 0, Y: 0, Z: -100},
					{X: 0, Y: 0, Z: 0},
				},
			},
		},
	}

	mesh := BuildPolyMesh(contourSet)
	if mesh == nil {
		t.Fail()
	}

	if len(mesh.Vertices) != 5 {
		t.Errorf("expected 5 vertices but found %d", len(mesh.Vertices))
	}
}

func TestTriangulate(t *testing.T) {
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
