package navmesh

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

type Span struct {
	Min, Max int
	Next     *Span
	Area     AREA_TYPE
	X, Z     int // purely for debugging, x and z can be removed
}

type HeightField struct {
	Width, Height int
	BMin, BMax    mgl64.Vec3

	Spans []*Span
}

func NewHeightField(width, height int, bMin, bMax mgl64.Vec3) *HeightField {
	return &HeightField{
		Width:  width,
		Height: height,
		BMin:   bMin,
		BMax:   bMax,
		Spans:  make([]*Span, width*height),
	}
}

func (hf *HeightField) SpanCount() int {
	count := 0
	for z := range hf.Height {
		for x := range hf.Width {
			index := x + z*hf.Width
			span := hf.Spans[index]
			for span != nil {
				if span.Area != NULL_AREA {
					count += 1
				}
				span = span.Next
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
	columnIndex := x + z*hf.Width
	currentSpan := hf.Spans[columnIndex]
	var previousSpan *Span

	// find where to insert
	for currentSpan != nil {
		if currentSpan.Min > y {
			break
		}

		if currentSpan.Max == y || currentSpan.Min == y {
			return
		}

		previousSpan = currentSpan
		currentSpan = currentSpan.Next
	}

	if previousSpan != nil {
		if previousSpan.Max == y-1 {
			// merge with prevous
			previousSpan.Max += 1
		} else {
			// add new span
			newSpan := &Span{Min: y, Max: y, Area: area, X: x, Z: z}
			newSpan.Next = previousSpan.Next
			previousSpan.Next = newSpan
			previousSpan = newSpan
		}
	} else {
		// no previous span means this is the lowest span
		newSpan := &Span{Min: y, Max: y, Area: area, X: x, Z: z}
		newSpan.Next = currentSpan
		hf.Spans[columnIndex] = newSpan
		previousSpan = newSpan
	}

	nextSpan := previousSpan.Next
	if nextSpan != nil {
		// merge with next span
		if previousSpan.Max == nextSpan.Min-1 {
			previousSpan.Max = nextSpan.Max
			previousSpan.Next = nextSpan.Next
		}
	}
}

func (hf *HeightField) AddSpan(x, z, sMin, sMax int, walkable bool, areaMergeThreshold int) {
	var previousSpan *Span
	columnIndex := x + z*hf.Width
	currentSpan := hf.Spans[columnIndex]

	area := NULL_AREA
	if walkable {
		area = WALKABLE_AREA
	}

	newSpan := &Span{Min: sMin, Max: sMax, Area: area, X: x, Z: z}

	for currentSpan != nil {
		if currentSpan.Min > newSpan.Max {
			break
		}

		if currentSpan.Max < newSpan.Min {
			previousSpan = currentSpan
			currentSpan = currentSpan.Next
		} else {
			// the current recast implementation favors the area of the new span over the new
			// if the max of each span is outside of the areaMergeThreshold
			if currentSpan.Min < newSpan.Min {
				newSpan.Min = currentSpan.Min
			}
			if currentSpan.Max > newSpan.Max {
				newSpan.Max = currentSpan.Max
			}

			// merge flags
			if int(math.Abs(float64(newSpan.Max-currentSpan.Max))) <= areaMergeThreshold {
				// higher area ID numbers indicate higher resolution priority
				// NULL_AREA is the smallest
				// WALKABLE_AREA is the largest
				newSpan.Area = max(newSpan.Area, currentSpan.Area)
			}

			next := currentSpan.Next
			if previousSpan != nil {
				previousSpan.Next = next
			} else {
				hf.Spans[columnIndex] = next
			}
			currentSpan = next
		}
	}

	if previousSpan != nil {
		newSpan.Next = previousSpan.Next
		previousSpan.Next = newSpan
	} else {
		newSpan.Next = hf.Spans[columnIndex]
		hf.Spans[columnIndex] = newSpan
	}
}

func (hf *HeightField) Test() {
	for z := range hf.Height {
		for x := range hf.Width {
			index := x + z*hf.Width
			span := hf.Spans[index]
			for span != nil {
				span = span.Next
			}
		}
	}
}
