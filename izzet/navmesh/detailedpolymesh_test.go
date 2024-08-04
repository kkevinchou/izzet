package navmesh

import (
	"testing"
)

func TestPolyMeshDetail(t *testing.T) {
	contourSet := &ContourSet{
		Contours: []Contour{
			{
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

	chf := &CompactHeightField{width: 200, height: 200}
	BuildDetailedPolyMesh(mesh, chf, nil)
}

func TestTriDist(t *testing.T) {
	verts := []DetailedVertex{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: -1},
	}
	tris := []Triangle{Triangle{Vertices: [3]int{0, 1, 2}}}
	vert := DetailedVertex{0, 5, 0}

	d := distToTris(vert, verts, tris)
	if d != 5 {
		t.Errorf("expected distance to be 5.0 but instead got %.1f", d)
		return
	}

	vert = DetailedVertex{0, -5, 0}
	d = distToTris(vert, verts, tris)
	if d != 5 {
		t.Errorf("expected distance to be 5.0 but instead got %.1f", d)
		return
	}
}

func TestCrossDirection(t *testing.T) {
	v0 := DetailedVertex{X: 640, Y: 301, Z: 550}
	v1 := DetailedVertex{X: 640, Y: 301, Z: 450}
	v2 := DetailedVertex{X: 600, Y: 301, Z: 500}

	res := vCross2D(v0, v1, v2)
	if res <= 0 {
		t.Errorf("crossing these two vectors should be positive (to the left)")
		return
	}
}
