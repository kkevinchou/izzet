package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
)

type CompactCell struct {
	SpanIndex SpanIndex
	SpanCount int
}

type CompactSpan struct {
	y         int
	regionID  int
	neighbors [4]SpanIndex

	h int
}

func (s CompactSpan) Y() int {
	return s.y
}

func (s CompactSpan) RegionID() int {
	return s.regionID
}

type CompactHeightField struct {
	width          int
	height         int
	spanCount      int
	walkableHeight int
	walkableClimb  int
	bMin, bMax     mgl64.Vec3
	cells          []CompactCell
	spans          []CompactSpan
	Distances      []int
	areas          []AREA_TYPE
	maxDistance    int
	maxRegionID    int
}

func NewCompactHeightField(walkableHeight, walkableClimb int, hf *HeightField) *CompactHeightField {
	chf := &CompactHeightField{}

	chf.width = hf.Width
	chf.height = hf.Height
	chf.spanCount = hf.SpanCount()
	chf.walkableHeight = walkableHeight
	chf.walkableClimb = walkableClimb
	chf.bMin = hf.BMin
	chf.bMax = hf.BMax
	chf.bMax[1] += float64(walkableHeight)
	chf.cells = make([]CompactCell, chf.width*chf.height)
	chf.spans = make([]CompactSpan, chf.spanCount)
	chf.areas = make([]AREA_TYPE, chf.spanCount)

	// initialize initial neighbor indices as negative to signal abscence of neighbors
	for i := 0; i < chf.spanCount; i++ {
		chf.spans[i].neighbors = [4]SpanIndex{-1, -1, -1, -1}
	}

	var currentSpanIndex SpanIndex
	numColumns := chf.width * chf.height

	for columnIndex := range numColumns {
		span := hf.Spans[columnIndex]
		if span == nil {
			continue
		}

		cell := &chf.cells[columnIndex]
		cell.SpanIndex = currentSpanIndex

		for ; span != nil; span = span.Next {
			if span.Area != NULL_AREA {
				bottom := span.Max
				top := maxHeight
				if span.Next != nil {
					top = span.Next.Min
				}
				chf.spans[currentSpanIndex].y = Clamp(bottom, 0, maxHeight)
				chf.spans[currentSpanIndex].h = Clamp(top-bottom, 0, maxHeight)
				chf.areas[currentSpanIndex] = span.Area
				// set areas
				currentSpanIndex++
				cell.SpanCount++
			}
		}
	}

	// setup neighbors

	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := &chf.spans[i]
				var count int
				for _, dir := range dirs {
					neighborX := x + xDirs[dir]
					neighborZ := z + zDirs[dir]

					if neighborX < 0 || neighborZ < 0 || neighborX >= chf.width || neighborZ >= chf.height {
						continue
					}

					neighborCell := &chf.cells[neighborX+neighborZ*chf.width]
					neighborSpanIndex := neighborCell.SpanIndex
					neighborSpanCount := neighborCell.SpanCount

					for j := neighborSpanIndex; j < neighborSpanIndex+SpanIndex(neighborSpanCount); j++ {
						neighborSpan := &chf.spans[j]

						bottom := Max(span.y, neighborSpan.y)
						top := Min(span.y+span.h, neighborSpan.y+neighborSpan.h)

						if top-bottom >= walkableHeight && Abs(neighborSpan.y-span.y) <= walkableClimb {
							layerIndex := int(j - neighborCell.SpanIndex)
							if layerIndex < 0 || layerIndex > maxLayers {
								panic("too many layers. could do a continue here to keep processing")
								// continue
							}
							count++
							span.neighbors[dir] = j
							break
						}
					}
				}
			}
		}
	}

	return chf
}

func (chf *CompactHeightField) Width() int {
	return chf.width
}

func (chf *CompactHeightField) Height() int {
	return chf.height
}

func (chf *CompactHeightField) BMin() mgl64.Vec3 {
	return chf.bMin
}

func (chf *CompactHeightField) BMax() mgl64.Vec3 {
	return chf.bMax
}

func (chf *CompactHeightField) Cells() []CompactCell {
	return chf.cells
}

func (chf *CompactHeightField) Spans() []CompactSpan {
	return chf.spans
}

func (chf *CompactHeightField) Test() {
	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
			}
		}
	}
}
