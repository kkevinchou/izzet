package navmesh

import "testing"

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
	BuildDetailedPolyMesh(mesh, chf)
}
