package navmesh

import "github.com/go-gl/mathgl/mgl64"

type CompactCell struct {
	SpanIndex int
	SpanCount int
}

type CompactSpan struct {
	y         int
	regionID  int
	neighbors [4]int

	h int
}

func (s CompactSpan) Y() int {
	return s.y
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
	distances      []int
	maxDistance    int
}

func NewCompactHeightField(walkableHeight, walkableClimb int, hf *HeightField) *CompactHeightField {
	chf := &CompactHeightField{}

	chf.width = hf.width
	chf.height = hf.height
	chf.spanCount = hf.SpanCount()
	chf.walkableHeight = walkableHeight
	chf.walkableClimb = walkableClimb
	chf.bMin = hf.bMin
	chf.bMax = hf.bMax
	chf.bMax[1] += float64(walkableHeight)
	chf.cells = make([]CompactCell, chf.width*chf.height)
	chf.spans = make([]CompactSpan, chf.spanCount)

	// initialize initial neighbor indices as negative to signal abscence of neighbors
	for i := 0; i < chf.spanCount; i++ {
		chf.spans[i].neighbors = [4]int{-1, -1, -1, -1}
	}

	currentSpanIndex := 0
	numColumns := chf.width * chf.height

	for columnIndex := range numColumns {
		span := hf.spans[columnIndex]
		if span == nil {
			continue
		}

		cell := &chf.cells[columnIndex]
		cell.SpanIndex = currentSpanIndex

		for ; span != nil && span.Valid(); span = span.next {
			bottom := span.max
			top := maxHeight
			if span.next != nil {
				top = span.next.min
			}
			chf.spans[currentSpanIndex].y = Clamp(bottom, 0, maxHeight)
			chf.spans[currentSpanIndex].h = Clamp(top-bottom, 0, maxHeight)
			currentSpanIndex++
			cell.SpanCount++
		}
	}

	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+spanCount; i++ {
				span := &chf.spans[spanIndex]
				for _, dir := range dirs {
					neighborX := x + xDirs[dir]
					neighborZ := z + zDirs[dir]

					if neighborX < 0 || neighborZ < 0 || neighborX >= chf.width || neighborZ >= chf.height {
						continue
					}

					neighborCell := &chf.cells[neighborX+neighborZ*chf.width]
					neighborSpanIndex := neighborCell.SpanIndex
					neighborSpanCount := neighborCell.SpanCount

					for j := neighborSpanIndex; j < neighborSpanIndex+neighborSpanCount; j++ {
						neighborSpan := &chf.spans[j]

						bottom := Max(span.y, neighborSpan.y)
						top := Min(span.y+span.h, neighborSpan.y+neighborSpan.h)

						if top-bottom >= walkableHeight && Abs(neighborSpan.y-span.y) <= walkableClimb {
							layerIndex := j - neighborCell.SpanIndex
							if layerIndex < 0 || layerIndex > maxLayers {
								panic("too many layers. could do a continue here to keep processing")
								// continue
							}
							span.neighbors[dir] = j
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
