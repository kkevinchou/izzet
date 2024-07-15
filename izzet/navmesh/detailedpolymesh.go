package navmesh

import (
	"fmt"
	"math"

	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
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

type Triangle struct {
	A, B, C int
}

func BuildDetailedPolyMesh(mesh *Mesh, chf *CompactHeightField, runtimeConfig *runtimeconfig.RuntimeConfig) *DetailedMesh {
	bounds := make([]Bound, len(mesh.Polygons))
	var maxhw, maxhh int
	var nPolyVerts int

	// TODO: this should be dynamic and come from another datastructure (chf, or mesh, or w/e)
	var cs float64 = 1
	var ch float64 = 1
	heightSearchRadius := max(1, int(math.Ceil(mesh.maxEdgeError)))

	// find max size for a polygon area
	for i := range mesh.Polygons {
		polygon := &mesh.Polygons[i]

		xmin := chf.width
		xmax := 0
		zmin := chf.height
		zmax := 0

		for _, vertIndex := range polygon.Verts {
			vert := mesh.Vertices[vertIndex]
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

	debugMap := runtimeConfig.DebugBlob2IntMap

	for i := range mesh.Polygons {
		if !debugMap[i] && len(debugMap) > 0 {
			continue
		}
		var polyVerts []DetailedVertex
		polygon := &mesh.Polygons[i]
		for _, vertIndex := range polygon.Verts {
			vert := mesh.Vertices[vertIndex]
			polyVerts = append(polyVerts, DetailedVertex{
				X: float64(vert.X) * cs, Y: float64(vert.Y) * ch, Z: float64(vert.Z) * cs,
			})
		}

		hp.xmin = bounds[i].xmin
		hp.zmin = bounds[i].zmin
		hp.width = bounds[i].xmax - bounds[i].xmin
		hp.height = bounds[i].zmax - bounds[i].zmin
		getHeightData(chf, hp, polygon.RegionID, i)
		buildDetailedPoly(chf, polyVerts, 1, 1, heightSearchRadius, hp, i)
		dmesh.Outlines = append(dmesh.Outlines, polyVerts)
	}

	return &dmesh
}

var HP map[string]bool

func getHeightData(chf *CompactHeightField, hp HeightPatch, regionID int, polyID int) {
	if HP == nil {
		HP = map[string]bool{}
	}
	for i := range len(hp.data) {
		hp.data[i] = hpUnsetHeight
	}

	var queue []QueueItem
	empty := true

	// copy the height from the same region, and mark region borders as seeds
	// to fill in the rest
	for hz := range hp.height {
		z := hp.zmin + hz
		for hx := range hp.width {
			x := hp.xmin + hx
			cell := chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := chf.spans[i]
				if span.regionID == regionID {
					hp.data[hx+hz*hp.width] = span.y
					HP[fmt.Sprintf("%d_%d", x, z)] = true
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
						queue = append(queue, QueueItem{x: x, z: z, spanIndex: i})
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

func buildDetailedPoly(chf *CompactHeightField, inVerts []DetailedVertex, sampleDist, sampleMaxError float64, heightSearchRadius int, hp HeightPatch, polygonID int) ([]DetailedVertex, []Triangle) {
	var ch float64 = 1
	var cs float64 = 1
	ics := 1 / cs
	var hull []int
	// heightSearchRadius = 100

	if polygonID == 4 {
		fmt.Println("HI")
	}

	verts := make([]DetailedVertex, len(inVerts))
	for i := range len(inVerts) {
		verts[i] = inVerts[i]
	}

	minExtent := minPolyExtent(inVerts)

	// tesselate outlines
	// this is done in a separate pass to ensure seamless height values across the poly boundaries
	for i, j := 0, len(inVerts)-1; i < len(inVerts); j, i = i, i+1 {
		vj := inVerts[j]
		vi := inVerts[i]
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
		if len(inVerts)+nn >= maxVerts {
			nn = maxVerts - 1 - len(inVerts)
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
				d := distancePtSeg(edges[m], v0, v1)
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
				hull = append(hull, len(verts))
				verts = append(verts, edges[idx[k]])
			}
		} else {
			for k := 1; k < nidx-1; k++ {
				hull = append(hull, len(verts))
				verts = append(verts, edges[idx[k]])
			}
		}
	}

	var tris []Triangle
	if minExtent < sampleDist*2 {
		tris = triangulateHull(verts, hull, len(inVerts))
		// setTriFlags
		return verts, tris
	}

	// tessellate the base mesh
	tris = triangulateHull(verts, hull, len(inVerts))
	return verts, tris
}

func triangulateHull(verts []DetailedVertex, hull []int, originalNumVerts int) []Triangle {
	// start from an ear with the shortest perimeter.
	// this tends to favor well formed triangles as a starting point

	start, left, right := 0, 1, len(hull)-1
	dmin := math.MaxFloat64

	for i := range len(hull) {
		if hull[i] >= originalNumVerts {
			// ears are centered on original hull vertices and not on newly added vertices from tesselation
			continue
		}

		pi := prev(i, len(hull))
		ni := next(i, len(hull))
		pv := verts[hull[pi]]
		cv := verts[hull[i]]
		nv := verts[hull[ni]]
		d := dist2D(pv, cv) + dist2D(cv, nv) + dist2D(nv, pv)
		if d < dmin {
			start = i
			left = ni
			right = pi
			dmin = d
		}
	}

	var tris []Triangle
	tris = append(tris, Triangle{A: hull[start], B: hull[left], C: hull[right]})

	// triangulate the polygon by moving left or right, depending on which triangle
	// has the shorter perimeter
	for next(left, len(hull)) != right {
		nLeft := next(left, len(hull))
		nRight := prev(right, len(hull))

		leftVert := verts[hull[left]]
		nextLeftVert := verts[hull[nLeft]]
		rightVert := verts[hull[right]]
		nextRightVert := verts[hull[nRight]]

		dLeft := dist2D(leftVert, nextLeftVert) + dist2D(nextLeftVert, rightVert)
		dRight := dist2D(rightVert, nextRightVert) + dist2D(leftVert, nextRightVert)

		if dLeft < dRight {
			tris = append(tris, Triangle{A: hull[left], B: hull[nLeft], C: hull[right]})
			left = nLeft
		} else {
			tris = append(tris, Triangle{A: hull[left], B: hull[right], C: hull[nRight]})
			right = nRight
		}
	}

	return tris
}

func next(i, max int) int {
	return (i + 1) % max
}
func prev(i, max int) int {
	return (i - 1 + max) % max
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
		// fmt.Printf("failed to sample %.0f %.0f %.0f\n", fx, fy, fz)
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
			d := distancePtSeg2Df(verts[i].X, verts[i].Z, p1.X, p1.Z, p2.X, p2.Z)
			maxEdgeDist = max(maxEdgeDist, d)
		}
		minDist = min(minDist, maxEdgeDist)
	}
	return math.Sqrt(minDist)
}

func distancePtSeg(pt, p, q DetailedVertex) float64 {
	pqx := q.X - p.X
	pqy := q.Y - p.Y
	pqz := q.Z - p.Z
	dx := pt.X - p.X
	dy := pt.Y - p.Y
	dz := pt.Z - p.Z
	d := pqx*pqx + pqy*pqy + pqz*pqz
	t := pqx*dx + pqy*dy + pqz*dz

	if d > 0 {
		t /= d
	}
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	dx = p.X + t*pqx - pt.X
	dy = p.Y + t*pqy - pt.Y
	dz = p.Z + t*pqz - pt.Z

	return dx*dx + dy*dy + dz*dz
}

// returns the squared distance between a point and a line segment
func distancePtSeg2Df(x, z, px, pz, qx, qz float64) float64 {
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

func dist2D(v0, v1 DetailedVertex) float64 {
	return math.Sqrt(dist2dSq(v0, v1))
}

func dist2dSq(v0, v1 DetailedVertex) float64 {
	dx := v0.X - v1.X
	dz := v0.Z - v1.Z
	return dx*dx + dz*dz
}
