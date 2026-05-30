package navmesh

import (
	"testing"

	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
)

func TestPolyMeshDetail(t *testing.T) {
	contourSet := &ContourSet{
		CellSize:   1,
		CellHeight: 1,
		Contours: []Contour{
			{
				RegionID: 1,
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

	chf := newFlatCompactHeightField(200, 200, 101, 1)
	BuildDetailedPolyMesh(mesh, chf, runtimeconfig.DefaultRuntimeConfig())
}

func newFlatCompactHeightField(width, height, populatedWidth, regionID int) *CompactHeightField {
	chf := &CompactHeightField{
		width:      width,
		height:     height,
		spanCount:  populatedWidth,
		cells:      make([]CompactCell, width*height),
		spans:      make([]CompactSpan, populatedWidth),
		CellSize:   1,
		CellHeight: 1,
	}

	for x := range populatedWidth {
		chf.cells[x] = CompactCell{SpanIndex: SpanIndex(x), SpanCount: 1}
		chf.spans[x] = CompactSpan{
			regionID:  regionID,
			neighbors: [4]SpanIndex{-1, -1, -1, -1},
		}
	}

	return chf
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
	v0 := DetailedVertex{X: 0, Y: 0, Z: 0}
	v1 := DetailedVertex{X: 1, Y: 0, Z: -1}
	v2 := DetailedVertex{X: 0, Y: 0, Z: -1}

	res := vCross2D(v0, v1, v2)
	if res >= 0 {
		t.Fatal("v0->v2 should be to the left of v0->v1, and therefore negative")
	}
}
