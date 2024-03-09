package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/constraints"
)

type Span struct {
	min, max int
	next     *Span
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

func (hf *HeightField) Spans() []*Span {
	return hf.spans
}

func (hf *HeightField) SpanCount() int {
	count := 0
	for _, span := range hf.spans {
		for span != nil {
			count += 1
			span = span.next
		}
	}
	return count
}

// AddVoxel adds a voxel to the heightfield, either extending an existing span, or creating a new one
func (hf *HeightField) AddVoxel(x, y, z int) {
	columnIndex := x + z*hf.width
	currentSpan := hf.spans[columnIndex]
	var previousSpan *Span

	// newSpan := &Span{min: min, max: max}

	for currentSpan != nil {
		if currentSpan.min > y {
			if currentSpan.min == y+1 {
				// attach to the current span
				currentSpan.min -= 1
			} else {
				// create a new span
				newSpan := &Span{min: y, max: y}
				newSpan.next = currentSpan
				hf.spans[columnIndex] = newSpan
			}
			return
		}

		if currentSpan.max < y {
			// attach to the current span
			if currentSpan.max == y-1 {
				currentSpan.max += 1
				if currentSpan.next != nil && currentSpan.next.min == currentSpan.max+1 {
					// merge agane
					currentSpan.max = currentSpan.next.max
					currentSpan.next = nil
				}
				return
			}
		}

		previousSpan = currentSpan
		currentSpan = currentSpan.next
	}

	if previousSpan != nil {
		newSpan := &Span{min: y, max: y}
		previousSpan.next = newSpan
	} else if currentSpan == nil {
		newSpan := &Span{min: y, max: y}
		hf.spans[columnIndex] = newSpan
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

func zmin[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

func zmax[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}
