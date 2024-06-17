package navmesh

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
)

// / Border vertex flag.
// / If a region ID has this bit set, then the associated element lies on
// / a tile border. If a contour vertex's region ID has this bit set, the
// / vertex will later be removed in order to match the segments and vertices
// / at tile boundaries.
// / (Used during the build process.)
const borderVertexFlag int = 0x10000

// / Area border flag.
// / If a region ID has this bit set, then the associated element lies on
// / the border of an area.
// / (Used during the region and contour build process.)
const areaBorderFlag int = 0x20000

type Vertex struct {
	X, Y, Z  int
	regionID int
	flags    int
}

type Contour struct {
	//
	//
	regionID  int
	area      AREA_TYPE
	Verts     []Vertex
	CellVerts []Vertex
	rawVerts  []Vertex
}

type ContourSet struct {
	bMin, bMax    mgl64.Vec3
	width, height int
	maxError      float64
	Contours      []Contour
}

func BuildContours(chf *CompactHeightField, maxError float64, maxEdgeLength int) *ContourSet {
	contourSet := &ContourSet{
		bMin:     chf.bMin,
		bMax:     chf.bMax,
		width:    chf.width,
		height:   chf.height,
		maxError: maxError,
		Contours: nil,
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

				regionID := chf.spans[i].regionID
				if regionID == 0 {
					continue
				}

				area := chf.areas[i]
				verts, cellVerts := getContourPoints(x, z, int(i), chf, flags)
				if len(verts) > 0 {
					fmt.Println("HI")
				}
				simplified := simplifyContour(verts, maxError, maxEdgeLength)
				simplified = removeDegenderateSegments(simplified)

				if len(simplified) >= 3 {
					contour := Contour{
						regionID:  regionID,
						area:      area,
						Verts:     simplified,
						CellVerts: cellVerts,
						rawVerts:  verts,
					}

					contourSet.Contours = append(contourSet.Contours, contour)
				}
			}
		}
	}

	// Merge holes
	if len(contourSet.Contours) > 0 {

	}

	return contourSet
}

func getContourPoints(x, z, i int, chf *CompactHeightField, flags []int) ([]Vertex, []Vertex) {
	dir := 0
	for (flags[i] & (1 << dir)) == 0 {
		dir++
	}

	startDir := dir
	starti := i
	area := chf.areas[i]
	var vertices []Vertex
	var cellVertices []Vertex

	for {
		span := chf.spans[i]
		neighborSpanIndex := span.neighbors[dir]

		if flags[i]&(1<<dir) > 0 {
			var isAreaBorder bool
			px := x
			py, isBorderVertex := getCornerHeight(x, z, i, dir, chf)
			pz := z

			if dir == 0 {
				pz--
			} else if dir == 1 {
				px++
				pz--
			} else if dir == 2 {
				px++
			}
			regionID := 0
			if neighborSpanIndex != -1 {
				neighborSpan := chf.spans[neighborSpanIndex]
				regionID = neighborSpan.regionID
				if area != chf.areas[neighborSpanIndex] {
					isAreaBorder = true
				}
			}
			var vFlags int
			if isBorderVertex {
				vFlags |= borderVertexFlag
			}
			if isAreaBorder {
				vFlags |= areaBorderFlag
			}
			vertices = append(vertices, Vertex{
				X:        px,
				Y:        py,
				Z:        pz,
				regionID: regionID,
				flags:    vFlags,
			})
			cellVertices = append(cellVertices, Vertex{
				X:        x,
				Y:        py,
				Z:        z,
				regionID: regionID,
				flags:    vFlags,
			})

			flags[i] &= ^(1 << dir) // remove visited edge
			dir = (dir + 1) % 4     // rotate CW
		} else {
			if neighborSpanIndex == -1 {
				panic("should not happen")
			}

			x = x + xDirs[dir]
			z = z + zDirs[dir]
			i = int(neighborSpanIndex)
			dir = (dir + 3) % 4 // rotate CCW
		}
		if starti == i && startDir == dir {
			break
		}
	}

	return vertices, cellVertices
}

func getCornerHeight(x, z, i, dir int, chf *CompactHeightField) (int, bool) {
	span := &chf.spans[i]
	cellHeight := span.y
	nextDir := (dir + 1) % 4
	var isBorderVertex bool

	var regions [4]int
	regions[0] = span.regionID | (int(chf.areas[i]) << 16)

	// grab the max height from the surrounding cells.
	// there are four cells considered:
	//	1 - itself
	// 	2 - dir (e.g. left)
	//  3 - dir + next dir (e.g. left + up)
	//  4 - next dir (e.g. up)

	neighborSpanIndex := span.neighbors[dir]
	if neighborSpanIndex != -1 {
		neighborSpan := chf.spans[neighborSpanIndex]
		cellHeight = max(cellHeight, neighborSpan.y)
		regions[1] = neighborSpan.regionID | (int(chf.areas[neighborSpanIndex]) << 16)

		if neighborSpanIndex2 := neighborSpan.neighbors[nextDir]; neighborSpanIndex2 != -1 {
			neighborSpan2 := chf.spans[neighborSpanIndex2]
			cellHeight = max(cellHeight, neighborSpan2.y)
			regions[2] = neighborSpan2.regionID | (int(chf.areas[neighborSpanIndex2]) << 16)
		}
	}

	neighborSpanIndex = span.neighbors[nextDir]
	if neighborSpanIndex != -1 {
		neighborSpan := chf.spans[neighborSpanIndex]
		cellHeight = max(cellHeight, neighborSpan.y)
		regions[3] = neighborSpan.regionID | (int(chf.areas[neighborSpanIndex]) << 16)
		if neighborSpanIndex2 := neighborSpan.neighbors[dir]; neighborSpanIndex2 != -1 {
			neighborSpan2 := chf.spans[neighborSpanIndex2]
			cellHeight = max(cellHeight, neighborSpan2.y)
			regions[2] = neighborSpan2.regionID | (int(chf.areas[neighborSpanIndex2]) << 16)
		}
	}

	// i'm not sure what this code does or why it's important

	// Check if the vertex is special edge vertex, these vertices will be removed later.
	for j := range 4 {
		a := j
		b := (j + 1) % 4
		c := (j + 2) % 4
		d := (j + 3) % 4

		// The vertex is a border vertex there are two same exterior cells in a row,
		// followed by two interior cells and none of the regions are out of bounds.
		twoSameExts := (regions[a]&regions[b]&borderVertexFlag) != 0 && regions[a] == regions[b]
		twoInts := ((regions[c] | regions[d]) & borderVertexFlag) == 0
		IntsSameArea := (regions[c] >> 16) == (regions[d] >> 16)
		noZeroes := regions[a] != 0 && regions[b] != 0 && regions[c] != 0 && regions[d] != 0
		if twoSameExts && twoInts && IntsSameArea && noZeroes {
			isBorderVertex = true
			break
		}
	}

	return cellHeight, isBorderVertex
}

func simplifyContour(vertices []Vertex, maxError float64, maxEdgeLength int) []Vertex {
	return vertices
}

func removeDegenderateSegments(vertices []Vertex) []Vertex {
	return vertices
}
