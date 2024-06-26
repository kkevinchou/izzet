package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Span struct {
	min, max int
	next     *Span
	area     AREA_TYPE
	x, z     int
}

func (s *Span) Min() int {
	return s.min
}

func (s *Span) Max() int {
	return s.max
}

func (s *Span) Next() *Span {
	return s.next
}

type HeightField struct {
	width, height int
	bMin, bMax    mgl64.Vec3

	spans []*Span
}

func NewHeightField(width, height int, bMin, bMax mgl64.Vec3) *HeightField {
	return &HeightField{
		width:  width,
		height: height,
		bMin:   bMin,
		bMax:   bMax,
		spans:  make([]*Span, width*height),
	}
}

// public methods just used for rendering
func (hf *HeightField) BMin() mgl64.Vec3 {
	return hf.bMin
}

func (hf *HeightField) BMax() mgl64.Vec3 {
	return hf.bMax
}

func (hf *HeightField) Spans() []*Span {
	return hf.spans
}

func (hf *HeightField) Width() int {
	return hf.width
}

func (hf *HeightField) Height() int {
	return hf.height
}

func (hf *HeightField) SpanCount() int {
	count := 0
	for z := range hf.height {
		for x := range hf.width {
			index := x + z*hf.width
			span := hf.spans[index]
			for span != nil {
				if span.area != NULL_AREA {
					count += 1
				}
				span = span.next
			}
		}
	}
	return count
}

// AddVoxel adds a voxel to the heightfield, either extending an existing span, or creating a new one
func (hf *HeightField) AddVoxel(x, y, z int, walkable bool) {
	area := NULL_AREA
	if walkable {
		area = WALKABLE_AREA
	}
	columnIndex := x + z*hf.width
	currentSpan := hf.spans[columnIndex]
	var previousSpan *Span

	// find where to insert
	for currentSpan != nil {
		if currentSpan.min > y {
			break
		}

		if currentSpan.max == y || currentSpan.min == y {
			return
		}

		previousSpan = currentSpan
		currentSpan = currentSpan.next
	}

	if previousSpan != nil {
		if previousSpan.max == y-1 {
			// merge with prevous
			previousSpan.max += 1
		} else {
			// add new span
			newSpan := &Span{min: y, max: y, area: area, x: x, z: z}
			newSpan.next = previousSpan.next
			previousSpan.next = newSpan
			previousSpan = newSpan
		}
	} else {
		// no previous span means this is the lowest span
		newSpan := &Span{min: y, max: y, area: area, x: x, z: z}
		newSpan.next = currentSpan
		hf.spans[columnIndex] = newSpan
		previousSpan = newSpan
	}

	nextSpan := previousSpan.next
	if nextSpan != nil {
		// merge with next span
		if previousSpan.max == nextSpan.min-1 {
			previousSpan.max = nextSpan.max
			previousSpan.next = nextSpan.next
		}
	}
}

func (hf *HeightField) AddSpan(x, z, min, max int) {
	var previousSpan *Span
	columnIndex := x + z*hf.width
	currentSpan := hf.spans[columnIndex]

	newSpan := &Span{min: min, max: max}

	for currentSpan != nil {
		if currentSpan.min > newSpan.max {
			break
		}

		if currentSpan.max < newSpan.min {
			previousSpan = currentSpan
			currentSpan = currentSpan.next
		} else {
			if currentSpan.min < newSpan.min {
				newSpan.min = currentSpan.min
			}
			if currentSpan.max > newSpan.max {
				newSpan.max = currentSpan.max
			}

			next := currentSpan.next
			if previousSpan != nil {
				previousSpan.next = next
			} else {
				hf.spans[columnIndex] = next
			}
			currentSpan = next
		}
	}

	if previousSpan != nil {
		newSpan.next = previousSpan.next
		previousSpan.next = newSpan
	} else {
		newSpan.next = hf.spans[columnIndex]
		hf.spans[columnIndex] = newSpan
	}
}

func (hf *HeightField) Test() {
	maxX := -1
	maxZ := -1
	_, _ = maxX, maxZ

	for z := range hf.height {
		for x := range hf.width {
			index := x + z*hf.width
			span := hf.Spans()[index]
			for span != nil {
				if span.x > maxX {
					maxZ = span.z
				}
				maxX = max(maxX, span.x)

				// minZ = min(minZ, span.z)
				// maxZ = max(maxZ, span.z)
				span = span.next
			}
		}
	}
}
