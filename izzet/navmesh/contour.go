package navmesh

import (
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
	x, y, z  int
	regionID int
	flags    int
}

type Contour struct {
	//
	//
	regionID int
	area     AREA_TYPE
	verts    []Vertex
	rawVerts []Vertex
}

type ContourSet struct {
	bMin, bMax    mgl64.Vec3
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
		contours: nil,
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
				verts := getContourPoints(x, z, int(i), chf, flags)
				simplified := simplifyContour(verts, maxError, maxEdgeLength)
				simplified = removeDegenderateSegments(simplified)

				if len(simplified)/4 >= 3 {
					contour := Contour{
						regionID: regionID,
						area:     area,
						verts:    simplified,
						rawVerts: verts,
					}

					contourSet.contours = append(contourSet.contours, contour)
				}
			}
		}
	}

	// Merge holes
	if len(contourSet.contours) > 0 {

	}

	return contourSet
}

func getContourPoints(x, z, i int, chf *CompactHeightField, flags []int) []Vertex {
	dir := 0
	for (flags[i] & (1 << dir)) == 0 {
		dir++
	}

	startDir := dir
	starti := i
	area := chf.areas[i]
	var vertices []Vertex

	for !(starti == i && startDir == dir) {
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
				x:        px,
				y:        py,
				z:        pz,
				regionID: regionID,
				flags:    vFlags,
			})

			flags[i] &= ^(1 << dir) // removed visited edge
			dir = (dir + 1) % 4     // rotate CW
		} else {
			neighbori := -1
			if neighborSpanIndex != -1 {
				neighbori = int(neighborSpanIndex)
			} else {
				panic("should not happen")
			}
			x = x + xDirs[dir]
			z = z + zDirs[dir]
			i = neighbori
			dir = (dir + 3) % 4
		}
	}

	return vertices
}

func getCornerHeight(x, z, i, dir int, chf *CompactHeightField) (int, bool) {
	return -1, false
}

func simplifyContour(points []Vertex, maxError float64, maxEdgeLength int) []Vertex {
	return nil
}

func removeDegenderateSegments(simplified []Vertex) []Vertex {
	return nil
}
