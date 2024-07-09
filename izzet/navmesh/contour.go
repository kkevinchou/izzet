package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
)

// Border vertex flag.
// If a region ID has this bit set, then the associated element lies on
// a tile border. If a contour vertex's region ID has this bit set, the
// vertex will later be removed in order to match the segments and vertices
// at tile boundaries.
// (Used during the build process.)
const borderVertexFlag int = 0x10000

// Area border flag.
// If a region ID has this bit set, then the associated element lies on
// the border of an area.
// (Used during the region and contour build process.)
const areaBorderFlag int = 0x20000

type Vertex struct {
	X, Y, Z  int
	regionID int
	flags    int
}

type SimplifiedVertex struct {
	X, Y, Z int
	I       int
}

type Contour struct {
	RegionID int
	area     AREA_TYPE
	Verts    []SimplifiedVertex
	RawVerts []Vertex
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

				flags[i] = res ^ 0xf // flags tracks which directions are not connected
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

				// TODO - implement
				simplified = removeDegenderateSegments(simplified)

				if len(simplified) >= 3 {
					contour := Contour{
						RegionID: regionID,
						area:     area,
						Verts:    simplified,
						RawVerts: verts,
					}

					contourSet.Contours = append(contourSet.Contours, contour)
				}
			}
		}
	}

	// Merge holes
	if len(contourSet.Contours) > 0 {
		var numHoles int
		winding := make([]int, len(contourSet.Contours))

		for i := range len(contourSet.Contours) {
			contour := &contourSet.Contours[i]
			area := calcAreaOfPolygon2D(contour.Verts)
			winding[i] = 1
			if area < 0 {
				winding[i] = -1
				numHoles++
			}

		}

		// TODO - merge holes
		if numHoles > 0 {
			// panic("holes detected")
			// numRegions := chf.maxRegionID + 1
		}
	}

	return contourSet
}

func calcAreaOfPolygon2D(verts []SimplifiedVertex) int {
	area := 0
	j := len(verts) - 1
	for i := 0; i < len(verts); i++ {
		vi := verts[i]
		vj := verts[j]
		area += (vi.X * vj.Z) - (vi.Z * vj.X)

		j = i
	}
	return (area + 1) / 2
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

	for {
		span := chf.spans[i]
		neighborSpanIndex := span.neighbors[dir]

		if flags[i]&(1<<dir) > 0 {
			var isAreaBorder bool
			px := x
			py, isBorderVertex := getCornerHeight(x, z, i, dir, chf)
			pz := z

			if dir == 0 {
				pz++
			} else if dir == 1 {
				px++
				pz++
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

			flags[i] &= ^(1 << dir) // remove visited edge
			dir = (dir + 1) % 4     // rotate CCW
		} else {
			if neighborSpanIndex == -1 {
				panic("should not happen")
			}

			x = x + xDirs[dir]
			z = z + zDirs[dir]
			i = int(neighborSpanIndex)
			dir = (dir + 3) % 4 // rotate CW
		}
		if starti == i && startDir == dir {
			break
		}
	}

	return vertices
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

func simplifyContour(vertices []Vertex, maxError float64, maxEdgeLength int) []SimplifiedVertex {
	// 1. setup initial points

	var hasConnections bool
	for _, vertex := range vertices {
		if vertex.regionID != 0 {
			hasConnections = true
			break
		}
	}

	var simplified []SimplifiedVertex

	if hasConnections {
		// the contour has some portals to other regions.
		// add a new point to every location where the region changes
		for i := 0; i < len(vertices); i++ {
			nexti := (i + 1) % len(vertices)
			differentRegions := vertices[i].regionID != vertices[nexti].regionID
			differentAreaBorder := (vertices[i].flags | areaBorderFlag) != (vertices[nexti].flags | areaBorderFlag)
			if differentRegions || differentAreaBorder {
				simplified = append(simplified, SimplifiedVertex{
					X: vertices[i].X,
					Y: vertices[i].Y,
					Z: vertices[i].Z,
					I: i,
				})
			}
		}
	}

	if len(simplified) == 0 {
		// if there are no connections at all,
		// create some initial points for the simplification process
		// find the lower left and upper right vertices of the contour

		lowerLeftX := vertices[0].X
		lowerLeftY := vertices[0].Y
		lowerLeftZ := vertices[0].Z
		lowerLeftI := 0

		upperRightX := vertices[0].X
		upperRightY := vertices[0].Y
		upperRightZ := vertices[0].Z
		upperRightI := 0

		for i, vertex := range vertices {
			x := vertex.X
			y := vertex.Y
			z := vertex.Z

			// maybe use greater than?
			if x < lowerLeftX || (x == lowerLeftX && z < lowerLeftZ) {
				lowerLeftX = x
				lowerLeftY = y
				lowerLeftZ = z
				lowerLeftI = i
			}

			if x > upperRightX || (x == upperRightX && z > upperRightZ) {
				upperRightX = x
				upperRightY = y
				upperRightZ = z
				upperRightI = i
			}
		}
		simplified = append(simplified, SimplifiedVertex{
			X: lowerLeftX,
			Y: lowerLeftY,
			Z: lowerLeftZ,
			I: lowerLeftI,
		})
		simplified = append(simplified, SimplifiedVertex{
			X: upperRightX,
			Y: upperRightY,
			Z: upperRightZ,
			I: upperRightI,
		})
	}

	// 2. add points until all raw points are within error tolerance

	maxErrorSq := maxError * maxError

	numVerts := len(vertices)
	for i := 0; i < len(simplified); {
		nexti := (i + 1) % len(simplified)

		ax := simplified[i].X
		az := simplified[i].Z
		ai := simplified[i].I

		bx := simplified[nexti].X
		bz := simplified[nexti].Z
		bi := simplified[nexti].I

		// find the maximum deviation from the segment

		// set up variables to either traverse a -> b or b -> a
		// this keep sthe actual traversal code simple and agnostic of direction

		var maxd float64
		maxi := -1
		var ci, cinc, endi int

		if bx > ax || (bx == ax && bz > az) {
			cinc = 1
			ci = (ai + cinc) % numVerts
			endi = bi
		} else {
			cinc = numVerts - 1
			ci = (bi + cinc) % numVerts
			endi = ai
			ax, bx = bx, ax
			az, bz = bz, az
		}

		// tessellate only outer edges or edges between areas
		if vertices[ci].regionID == 0 || (vertices[ci].flags&areaBorderFlag > 0) {
			for ci != endi {
				d := distancePtSeg(vertices[ci].X, vertices[ci].Z, ax, az, bx, bz)
				if d > maxd {
					maxd = d
					maxi = ci
				}
				ci = (ci + cinc) % numVerts
			}
		}

		// if the max deviation is larger than the max error,
		// add a new point, otherwise move on to the next segment

		if maxi != -1 && maxd > maxErrorSq {
			simplified = append(simplified, SimplifiedVertex{
				X: vertices[maxi].X,
				Y: vertices[maxi].Y,
				Z: vertices[maxi].Z,
				I: maxi,
			})

			// bubble the new vertex down
			n := len(simplified)
			for j := n - 1; j > i+1; j-- {
				simplified[j], simplified[j-1] = simplified[j-1], simplified[j]
			}
		} else {
			i++
		}
	}

	return simplified
}

func distancePtSeg(x, z, px, pz, qx, qz int) float64 {
	pqx := float64(qx - px)
	pqz := float64(qz - pz)
	dx := float64(x - px)
	dz := float64(z - pz)
	d := pqx*pqx + pqz*pqz
	t := pqx*dx + pqz*dz

	if d > 0 {
		t /= d
	}
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	dx = float64(px) + t*pqx - float64(x)
	dz = float64(pz) + t*pqz - float64(z)

	return dx*dx + dz*dz
}

func removeDegenderateSegments(vertices []SimplifiedVertex) []SimplifiedVertex {
	return vertices
}
