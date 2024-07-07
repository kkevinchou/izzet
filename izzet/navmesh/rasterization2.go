package navmesh

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

type RasterVertex struct {
	X, Y, Z float64
}

type AxisType string

const (
	AxisTypeX AxisType = "X"
	AxisTypeZ AxisType = "Z"
)

func RasterizeTriangle2(v0, v1, v2 mgl64.Vec3, cellSize, inverseCellSize, inverseCellHeight float64, hf *HeightField, walkable bool, areaMergeThreshold int) int {
	triBBMin := v0
	vMin(&triBBMin, &v1)
	vMin(&triBBMin, &v2)

	triBBMax := v0
	vMax(&triBBMax, &v1)
	vMax(&triBBMax, &v2)

	if !overlapBounds(triBBMin, triBBMax, hf.bMin, hf.bMax) {
		return -1
	}

	w := hf.width
	h := hf.height
	by := hf.bMax.Y() - hf.bMin.Y()

	// calculate the footprint of the triangle on the grid's z axis
	z0 := int((triBBMin.Z() - hf.bMin.Z()) * inverseCellSize)
	z1 := int((triBBMax.Z() - hf.bMin.Z()) * inverseCellSize)

	// use -1 rather than 0 to cut the polygon properly at the start of the tile
	z0 = Clamp(z0, -1, h-1)
	z1 = Clamp(z1, 0, h-1)

	// not sure if this is actually faster than making 4 different slices
	// theoretically this should be more cache friendly since they're all packed together
	var buffer [7 * 5]RasterVertex
	in := buffer[0:7]
	zSlice := buffer[7:14]
	// technically we don't need a buffer for xSlice, we could reuse `in` in the x loop
	// however, this is more readable
	xSlice := buffer[14:21]
	zRemainder := buffer[21:28]
	xRemainder := buffer[28:35]

	in[0] = RasterVertex{X: v0.X(), Y: v0.Y(), Z: v0.Z()}
	in[1] = RasterVertex{X: v1.X(), Y: v1.Y(), Z: v1.Z()}
	in[2] = RasterVertex{X: v2.X(), Y: v2.Y(), Z: v2.Z()}

	var zSliceSize int
	var zRemainderSize int = 3

	for z := z0; z <= z1; z++ {
		cellZ := hf.bMin.Z() + float64(z)*cellSize
		zSliceSize, zRemainderSize = dividePoly(in, zRemainderSize, zSlice, zRemainder, cellZ+cellSize, AxisTypeZ)
		bufferSwap(&in, &zRemainder)

		if zSliceSize < 3 {
			continue
		}
		if z < 0 {
			continue
		}

		minX := zSlice[0].X
		maxX := zSlice[0].X

		for i := 0; i < zSliceSize; i++ {
			minX = min(minX, zSlice[i].X)
			maxX = max(maxX, zSlice[i].X)
		}

		x0 := int((minX - hf.bMin.X()) * inverseCellSize)
		x1 := int((maxX - hf.bMin.X()) * inverseCellSize)

		if x1 < 0 || x0 >= w {
			continue
		}

		x0 = Clamp(x0, -1, w-1)
		x1 = Clamp(x1, 0, w-1)

		var xSliceSize int
		xRemainderSize := zSliceSize

		for x := x0; x <= x1; x++ {
			cx := hf.bMin.X() + float64(x)*cellSize
			xSliceSize, xRemainderSize = dividePoly(zSlice, xRemainderSize, xSlice, xRemainder, cx+cellSize, AxisTypeX)
			bufferSwap(&zSlice, &xRemainder)

			if xSliceSize < 3 {
				continue
			}

			if x < 0 {
				continue
			}

			spanMin := xSlice[0].Y
			spanMax := xSlice[0].Y
			for i := 0; i < xSliceSize; i++ {
				spanMin = min(spanMin, xSlice[i].Y)
				spanMax = Max(spanMax, xSlice[i].Y)
			}
			spanMin -= hf.bMin.Y()
			spanMax -= hf.bMin.Y()

			// skip the span if it's outside the heightfield bounding box
			if spanMax < 0 {
				continue
			}

			if spanMin > by {
				continue
			}

			// clamp the span to the heightfield boundign box
			if spanMin < 0 {
				spanMin = 0
			}

			if spanMax > by {
				spanMax = by
			}

			// snap the span to the heightfield height grid
			spanMinCellIndex := Clamp(int(math.Floor(spanMin*inverseCellHeight)), 0, spanMaxHeight)
			spanMaxCellIndex := Clamp(int(math.Ceil(spanMax*inverseCellHeight)), spanMinCellIndex+1, spanMaxHeight)

			hf.AddSpan(x, z, spanMinCellIndex, spanMaxCellIndex, walkable, areaMergeThreshold)
		}
	}

	return -1
}

func dividePoly(inVerts []RasterVertex, inVertsCount int, outVerts1 []RasterVertex, outVerts2 []RasterVertex, axisOffset float64, axisType AxisType) (int, int) {
	if inVertsCount > 12 {
		panic(fmt.Sprintf("divide poly only supports splitting polygons of size up to 12, received %d", inVertsCount))
	}

	// how far positive or negative away from the separating axis for each vertex
	var inVertAxisDelta [12]float64
	for i := 0; i < inVertsCount; i++ {
		var axisValue float64
		if axisType == AxisTypeX {
			axisValue = inVerts[i].X
		} else if axisType == AxisTypeZ {
			axisValue = inVerts[i].Z
		} else {
			panic(fmt.Sprintf("unexpected axis type %s", axisType))
		}

		inVertAxisDelta[i] = axisOffset - axisValue
	}

	var poly1Vert int
	var poly2Vert int

	inVertB := inVertsCount - 1
	for inVertA := 0; inVertA < inVertsCount; inVertA++ {
		// check if the two verts are on the same side of the separating axis
		sameSide := (inVertAxisDelta[inVertA] >= 0) == (inVertAxisDelta[inVertB] >= 0)

		if !sameSide {
			s := inVertAxisDelta[inVertB] / (inVertAxisDelta[inVertB] - inVertAxisDelta[inVertA])

			// find the intersection with the separating axis and add that vertex to both buffers
			outVerts1[poly1Vert].X = inVerts[inVertB].X + (inVerts[inVertA].X-inVerts[inVertB].X)*s
			outVerts1[poly1Vert].Y = inVerts[inVertB].Y + (inVerts[inVertA].Y-inVerts[inVertB].Y)*s
			outVerts1[poly1Vert].Z = inVerts[inVertB].Z + (inVerts[inVertA].Z-inVerts[inVertB].Z)*s

			outVerts2[poly2Vert] = outVerts1[poly1Vert]

			poly1Vert++
			poly2Vert++

			// add inVertA to the right polygon. Do not add points that are on the dividing line
			// since these are added above

			if inVertAxisDelta[inVertA] > 0 {
				outVerts1[poly1Vert] = inVerts[inVertA]
				poly1Vert++
			} else if inVertAxisDelta[inVertA] < 0 {
				outVerts2[poly2Vert] = inVerts[inVertA]
				poly2Vert++
			}
		} else {
			// add inVertA to the right polygon. Addition is done even for points on the dividing line
			if inVertAxisDelta[inVertA] >= 0 {
				outVerts1[poly1Vert] = inVerts[inVertA]
				poly1Vert++
				if inVertAxisDelta[inVertA] != 0 {
					inVertB = inVertA
					continue
				}
			}
			outVerts2[poly2Vert] = inVerts[inVertA]
			poly2Vert++
		}
		inVertB = inVertA
	}

	return poly1Vert, poly2Vert
}

func bufferSwap(a, b *[]RasterVertex) {
	temp := *a
	*a = *b
	*b = temp
}

// vMin updates v0 to contain the minimum values for each axis between v0 and v1
func vMin(v0, v1 *mgl64.Vec3) {
	v0[0] = min(v0[0], v1[0])
	v0[1] = min(v0[1], v1[1])
	v0[2] = min(v0[2], v1[2])
}

// vMin updates v0 to contain the maximum values for each axis between v0 and v1
func vMax(v0, v1 *mgl64.Vec3) {
	v0[0] = max(v0[0], v1[0])
	v0[1] = max(v0[1], v1[1])
	v0[2] = max(v0[2], v1[2])
}

func overlapBounds(aMin, aMax, bMin, bMax mgl64.Vec3) bool {
	return aMin[0] <= bMax[0] && aMax[0] >= bMin[0] &&
		aMin[1] <= bMax[1] && aMax[1] >= bMin[1] &&
		aMin[2] <= bMax[2] && aMax[2] >= bMin[2]
}

// // vCopy copies a vec3 into a slice of floats
// func vCopy(v mgl64.Vec3) []float64 {
// 	r := make([]float64, 3)
// 	r[0] = v[0]
// 	r[1] = v[1]
// 	r[2] = v[2]
// 	return r
// }
