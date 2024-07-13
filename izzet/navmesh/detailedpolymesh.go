package navmesh

import (
	"math"
)

const (
	maxVertsPerEdge = 32
	maxVerts        = 127
	hpUnsetHeight   = 0xffff
)

type DetailedVertex struct {
	X, Y, Z float64
}

type HeightPatch struct {
	xmin   int
	zmin   int
	width  int
	height int

	data []int
}

type Bound struct {
	xmin int
	xmax int
	zmin int
	zmax int
}

type QueueItem struct {
	x, z      int
	spanIndex SpanIndex
}

type DetailedMesh struct {
	Outlines [][]DetailedVertex
}

func BuildDetailedPolyMesh(mesh *Mesh, chf *CompactHeightField) *DetailedMesh {
	bounds := make([]Bound, len(mesh.polygons))
	var maxhw, maxhh int
	var nPolyVerts int

	// TODO: this should be dynamic and come from another datastructure (chf, or mesh, or w/e)
	var cs float64 = 1
	var ch float64 = 1
	heightSearchRadius := max(1, int(math.Ceil(mesh.maxEdgeError)))

	// find max size for a polygon area
	for i := range mesh.polygons {
		polygon := &mesh.polygons[i]

		xmin := chf.width
		xmax := 0
		zmin := chf.height
		zmax := 0

		for _, vertIndex := range polygon.verts {
			vert := mesh.vertices[vertIndex]
			xmin = min(xmin, vert.X)
			xmax = max(xmax, vert.X)
			zmin = min(zmin, vert.Z)
			zmax = max(zmax, vert.Z)
			nPolyVerts++
		}

		xmin = max(0, xmin-1)
		xmax = min(chf.width, xmax+1)
		zmin = max(0, zmin-1)
		zmax = min(chf.height, zmax+1)

		bounds[i].xmin = xmin
		bounds[i].xmax = xmax
		bounds[i].zmin = zmin
		bounds[i].zmax = zmax

		if xmin >= xmax || zmin >= zmax {
			continue
		}

		maxhw = max(maxhw, xmax-xmin)
		maxhh = max(maxhh, zmax-zmin)
	}

	var hp HeightPatch
	hp.data = make([]int, maxhw*maxhh)

	var dmesh DetailedMesh

	for i := range mesh.polygons {
		var polyVerts []DetailedVertex
		polygon := &mesh.polygons[i]
		for _, vertIndex := range polygon.verts {
			vert := mesh.vertices[vertIndex]
			polyVerts = append(polyVerts, DetailedVertex{
				X: float64(vert.X) * cs, Y: float64(vert.Y) * ch, Z: float64(vert.Z) * cs,
			})
		}

		hp.xmin = bounds[i].xmin
		hp.zmin = bounds[i].zmin
		hp.width = bounds[i].xmax - bounds[i].xmin
		hp.height = bounds[i].zmax - bounds[i].zmin
		getHeightData(chf, hp, mesh.regionIDs[i])
		buildDetailedPoly(chf, &polyVerts, 1, 1, heightSearchRadius, hp)
		dmesh.Outlines = append(dmesh.Outlines, polyVerts)
	}

	return &dmesh
}

func getHeightData(chf *CompactHeightField, hp HeightPatch, regionID int) {
	for i := range len(hp.data) {
		hp.data[i] = -1
	}

	var queue []QueueItem
	empty := true

	// copy the height from the same region, and mark region borders as seeds
	// to fill in the rest
	for hz := range hp.height {
		z := hp.zmin + hz
		for hx := range hp.height {
			x := hp.xmin + hx
			cell := chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := chf.spans[spanIndex]
				if span.regionID == regionID {
					hp.data[hx+hz*hp.width] = span.y
					empty = false
					border := false
					for _, dir := range dirs {
						neighborSpanIndex := span.neighbors[dir]
						if neighborSpanIndex == -1 {
							continue
						}

						neighborSpan := chf.spans[neighborSpanIndex]
						if neighborSpan.regionID != regionID {
							border = true
							break
						}
					}
					if border {
						queue = append(queue, QueueItem{x: x, z: z, spanIndex: spanIndex})
					}
					break
				}
			}
		}
	}

	if empty {
		panic("TODO: implement seedArrayWithPolyCenter")
	}

	// assume the seed is centered in the polygon, so a BFS to collect
	// height data will ensure we do not move onto overlapping polygons and
	// sample wrong heights
	for len(queue) > 0 {
		queueItem := queue[0]
		queue = queue[1:]

		x := queueItem.x
		z := queueItem.z
		spanIndex := queueItem.spanIndex

		span := chf.spans[spanIndex]

		for _, dir := range dirs {
			neighborSpanIndex := span.neighbors[dir]
			if neighborSpanIndex == -1 {
				continue
			}
			neighborX := x + xDirs[dir]
			neighborZ := z + zDirs[dir]
			hx := neighborX - hp.xmin
			hy := neighborZ - hp.zmin

			if hx >= hp.width || hy >= hp.height || hx < 0 || hy < 0 {
				continue
			}

			if hp.data[hx+hy*hp.width] != hpUnsetHeight {
				continue
			}

			neighborSpan := chf.spans[neighborSpanIndex]

			hp.data[hx+hy*hp.width] = neighborSpan.y
			queue = append(queue, QueueItem{x: neighborX, z: neighborZ, spanIndex: neighborSpanIndex})
		}
	}
}

func buildDetailedPoly(chf *CompactHeightField, verts *[]DetailedVertex, sampleDist, sampleMaxError float64, heightSearchRadius int, hp HeightPatch) {
	var ch float64 = 1
	var cs float64 = 1
	ics := 1 / cs
	var hull []int

	minExtent := minPolyExtent(*verts)

	// tesselate outlines
	// this is done in a separate pass to ensure seamless height values across the poly boundaries
	for i, j := 0, len(*verts)-1; i < len(*verts); j, i = i, i+1 {
		vj := (*verts)[j]
		vi := (*verts)[i]
		swapped := false

		// make sure the segments are always handled in same order
		// using lexological sort or else there will be seams.
		if math.Abs(vj.X-vi.X) < 1e-6 {
			if vj.Z > vi.Z {
				vj, vi = vi, vj
				swapped = true
			}
		} else {
			if vj.X > vi.X {
				vj, vi = vi, vj
				swapped = true
			}
		}

		dx := vi.X - vj.X
		dy := vi.Y - vj.Y
		dz := vi.Z - vj.Z
		d := math.Sqrt(dx*dx + dz*dz)
		nn := 1 + int(math.Floor(d/sampleDist))

		if nn >= maxVertsPerEdge {
			nn = maxVertsPerEdge - 1
		}
		if len(*verts)+nn >= maxVerts {
			nn = maxVerts - 1 - len(*verts)
		}

		var edges []DetailedVertex
		for k := 0; k <= nn; k++ {
			u := float64(k) / float64(nn)
			x := vj.X + dx*u
			y := vj.Y + dy*u
			z := vj.Z + dz*u
			edges = append(edges, DetailedVertex{
				X: x,
				Y: float64(getHeight(x, y, z, cs, ics, ch, heightSearchRadius, hp)) * ch,
				Z: z,
			})
		}

		// simplify samples

		var idx [maxVertsPerEdge]int
		idx[0], idx[1] = 0, nn
		nidx := 2

		for k := 0; k < nidx-1; {
			i0 := idx[k]
			i1 := idx[k+1]
			v0 := edges[i0]
			v1 := edges[i1]
			var maxd float64
			maxi := -1
			for m := i0 + 1; m < i1; m++ {
				d := distancePtSegf(edges[m].X, edges[m].Z, v0.X, v0.Z, v1.X, v1.Z)
				if d > maxd {
					maxd = d
					maxi = m
				}
			}

			if maxi != -1 && maxd > (sampleMaxError*sampleMaxError) {
				for m := nidx; m > k; m-- {
					idx[m] = idx[m-1]
				}
				idx[k+1] = maxi
				nidx++
			} else {
				k++
			}
		}

		hull = append(hull, j)

		if swapped {
			for k := nidx - 2; k > 0; k-- {
				*verts = append(*verts, edges[idx[k]])
				hull = append(hull, len(*verts))
			}
		} else {
			for k := 1; k < nidx-1; k++ {
				*verts = append(*verts, edges[idx[k]])
				hull = append(hull, len(*verts))
			}
		}
	}

	if minExtent < sampleDist*2 {
		// triangulateHull
		// setTriFlags
		return
	}
}

func getHeight(fx, fy, fz, cs, ics, ch float64, radius int, hp HeightPatch) int {
	ix := int(math.Floor(fx*ics + 0.01))
	iz := int(math.Floor(fz*ics + 0.01))

	ix = Clamp(ix-hp.xmin, 0, hp.width-1)
	iz = Clamp(iz-hp.zmin, 0, hp.height-1)

	h := hp.data[ix+iz*hp.width]

	// spiral search for next available height value
	if h == hpUnsetHeight {
		x, z, dx, dz := 1, 0, 1, 0
		maxSize := radius*2 + 1
		maxIter := maxSize*maxSize - 1

		nextRingIterStart := 8
		nextRingIters := 16

		dmin := math.MaxFloat64
		for i := range maxIter {
			nx := ix + x
			nz := iz + z

			if nx >= 0 && nz >= 0 && nx < hp.width && nz < hp.height {
				nh := hp.data[nx+nz*hp.width]
				if nh != hpUnsetHeight {
					d := math.Abs(float64(nh)*ch - fy)
					if d < dmin {
						h = nh
						dmin = d
					}
				}
			}

			if i+1 == nextRingIterStart {
				if h != hpUnsetHeight {
					break
				}

				nextRingIterStart += nextRingIters
				nextRingIters += 8
			}

			if (x == z) || ((x < 0) && (x == -z)) || ((x > 0) && (x == 1-z)) {
				dx, dz = -dz, dx
			}
			x += dx
			z += dz
		}
	}

	if h == hpUnsetHeight {
		panic("failed to find height")
	}

	return h
}

func minPolyExtent(verts []DetailedVertex) float64 {
	var minDist float64 = math.MaxFloat64
	for i := range len(verts) {
		ni := (i + 1) % len(verts)
		p1 := verts[i]
		p2 := verts[ni]
		var maxEdgeDist float64 = 0
		for j := 0; j < len(verts); j++ {
			if j == i || j == ni {
				continue
			}
			d := distancePtSegf(verts[i].X, verts[i].Z, p1.X, p1.Z, p2.X, p2.Z)
			maxEdgeDist = max(maxEdgeDist, d)
		}
		minDist = min(minDist, maxEdgeDist)
	}
	return math.Sqrt(minDist)
}

// returns the squared distance between a point and a line segment
func distancePtSegf(x, z, px, pz, qx, qz float64) float64 {
	pqx := qx - px
	pqz := qz - pz
	dx := x - px
	dz := z - pz
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

	dx = px + t*pqx - x
	dz = pz + t*pqz - z

	return dx*dx + dz*dz
}
