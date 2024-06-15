package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Contour struct {
	//
	//
}

type ContourSet struct {
	bMin, bMax    mgl64.Vec3
	NumContours   int
	width, height int
	maxError      float64
	contours      []Contour
}

func BuildContours(chf *CompactHeightField, maxError float64, maxEdgeLength int) *ContourSet {
	contourSet := &ContourSet{
		bMin:     chf.bMin,
		bMax:     chf.bMax,
		width:    chf.width,
		height:   chf.height,
		maxError: maxError,
		contours: make([]Contour, chf.maxRegions),
	}

	flags := make([]int, chf.spanCount)

	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount
			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := &chf.spans[i]
				if span.regionID == 0 {
					continue
				}

				var res int // stores which direction is connected
				for _, dir := range dirs {
					var neighborRegionID int

					neighborSpanIndex := span.neighbors[dir]
					if neighborSpanIndex != -1 {
						neighborRegionID = chf.spans[neighborSpanIndex].regionID
					}

					if span.regionID == neighborRegionID {
						res |= (1 << dir)
					}
				}

				flags[i] = res ^ 0xf // flags tracks which directions are not connect
			}
		}
	}

	var verts [256]int
	var simpliied [64]int
	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount
			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				if flags[i] == 0 || flags[i] == 0xf {
					flags[i] = 0
					continue
				}
			}
		}
	}

	_, _ = verts, simpliied

	return contourSet
}
