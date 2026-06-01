package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestProjectPathPointUsesExplicitPolygon(t *testing.T) {
	tile := CTile{
		Vertices: []mgl64.Vec3{
			{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1},
			{2, 3, 0}, {2, 3, 1},
		},
		Polygons: []CPolygon{
			{Vertices: []int{0, 1, 2, 3}},
			{Vertices: []int{1, 4, 5, 2}},
		},
		DetailedPolygon: []CDetailedPolygon{
			{},
			{Triangles: []CDetailedTriangle{
				{Vertices: [3]int{0, 1, 2}},
				{Vertices: [3]int{0, 2, 3}},
			}},
		},
		DetailedVertices: [][]mgl64.Vec3{
			nil,
			{
				{1, 3, 0}, {2, 3, 0}, {2, 3, 1}, {1, 3, 1},
			},
		},
	}

	point := projectPathPoint(tile, 1, mgl64.Vec3{1.5, 0, 0.5})

	if got, want := point.Y(), 3.0; got != want {
		t.Fatalf("projected y = %v, want %v", got, want)
	}
}
