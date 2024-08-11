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
	edgeUndefined   = -1
	edgeHull        = -2
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
	PolyVertices    [][]DetailedVertex
	PolyTriangles   [][]Triangle
	OutlineSamples  [][]float32
	InteriorSamples [][]float32
	AllSamples      [][]float32
}

type Triangle struct {
	// A, B, C  int
	Vertices [3]int
	OnHull   [3]bool
}

type Sample struct {
	X, Y, Z int
	added   bool
}

type DetailedEdge struct {
	s, t int
	l, r int
}

func BuildDetailedPolyMesh(mesh *Mesh, chf *CompactHeightField, runtimeConfig *runtimeconfig.RuntimeConfig) *DetailedMesh {
	bounds := make([]Bound, len(mesh.Polygons))
	var maxhw, maxhh int
	var nPolyVerts int

	var cs float64 = mesh.CellSize
	var ch float64 = mesh.CellHeight

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

	heightSearchRadius := max(1, int(math.Ceil(mesh.maxEdgeError)))
	origin := mesh.bMin

	dmesh.PolyVertices = make([][]DetailedVertex, len(mesh.Polygons))
	dmesh.PolyTriangles = make([][]Triangle, len(mesh.Polygons))
	dmesh.OutlineSamples = make([][]float32, len(mesh.Polygons))
	dmesh.InteriorSamples = make([][]float32, len(mesh.Polygons))
	dmesh.AllSamples = make([][]float32, len(mesh.Polygons))

	for i := range mesh.Polygons {
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
		getHeightData(chf, hp, polygon.RegionID)
		verts, tris := buildDetailedPoly(chf, polyVerts, float64(runtimeConfig.NavigationmeshSampleDist), float64(runtimeConfig.NavigationmeshMaxError), heightSearchRadius, hp, &dmesh, i, runtimeConfig)

		// move detailed verts to world space
		for j := range len(verts) {
			v := &verts[j]
			v.X += origin.X()
			// v.Y += origin.Y() + chf.CellHeight
			v.Y += origin.Y()
			v.Z += origin.Z()
		}

		// offset poly too for flag checking
		for j := range len(polyVerts) {
			p := &polyVerts[j]
			p.X += origin.X()
			p.Y += origin.Y()
			p.Z += origin.Z()
		}

		dmesh.PolyVertices[i] = verts
		dmesh.PolyTriangles[i] = tris
	}

	return &dmesh
}

// var HP map[string]bool

func getHeightData(chf *CompactHeightField, hp HeightPatch, regionID int) {
	// if HP == nil {
	// 	HP = map[string]bool{}
	// }
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
					// HP[fmt.Sprintf("%d_%d", x, z)] = true
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

func buildDetailedPoly(chf *CompactHeightField, inVerts []DetailedVertex, sampleDist, sampleMaxError float64, heightSearchRadius int, hp HeightPatch, dmesh *DetailedMesh, polyIndex int, runtimeConfig *runtimeconfig.RuntimeConfig) ([]DetailedVertex, []Triangle) {
	cs := chf.CellSize
	ch := chf.CellHeight
	ics := 1 / cs
	var hull []int

	verts := make([]DetailedVertex, len(inVerts))
	for i := range len(inVerts) {
		verts[i] = inVerts[i]
	}

	minExtent := minPolyExtent(verts)

	// tesselate outlines
	// this is done in a separate pass to ensure seamless height values across the poly boundaries
	if sampleDist > 0 {
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
					Y: float64(getHeight(x, y, z, ics, ch, heightSearchRadius, hp)) * ch,
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
	}

	for _, i := range hull {
		dmesh.OutlineSamples[polyIndex] = append(dmesh.OutlineSamples[polyIndex], float32(verts[i].X), float32(verts[i].Y), float32(verts[i].Z))
	}

	var tris []Triangle

	if minExtent < sampleDist*2 {
		tris = triangulateHull(verts, hull, len(inVerts))
		// setTriFlags
		return verts, tris
	}
	_ = minExtent

	// tessellate the base mesh
	tris = append(tris, triangulateHull(verts, hull, len(inVerts))...)
	if len(tris) == 0 {
		panic("no triangles found from hull")
	}

	if sampleDist > 0 {
		bmin, bmax := inVerts[0], inVerts[0]
		for i := 1; i < len(inVerts); i++ {
			dvMin(&bmin, &inVerts[i])
			dvMax(&bmax, &inVerts[i])
		}

		x0 := int(math.Floor(bmin.X / sampleDist))
		x1 := int(math.Ceil(bmax.X / sampleDist))
		z0 := int(math.Floor(bmin.Z / sampleDist))
		z1 := int(math.Ceil(bmax.Z / sampleDist))

		var samples []Sample
		for z := z0; z < z1; z++ {
			for x := x0; x < x1; x++ {
				px := float64(x) * sampleDist
				py := (bmax.Y + bmin.Y) / 2
				pz := float64(z) * sampleDist

				// make sure the samples are not too close to the edges
				if distToPoly(inVerts, DetailedVertex{X: px, Y: py, Z: pz}) > -sampleDist/2 {
					continue
				}
				sample := Sample{
					X: x,
					Y: getHeight(px, py, pz, ics, ch, heightSearchRadius, hp),
					Z: z,
				}
				samples = append(samples, sample)

				sx := float64(sample.X) * sampleDist
				sy := float64(sample.Y) * ch
				sz := float64(sample.Z) * sampleDist
				dmesh.AllSamples[polyIndex] = append(dmesh.AllSamples[polyIndex], float32(sx), float32(sy), float32(sz))
			}
		}

		// progressively add samples which have the highest error until
		// we go under the error threshold
		for iter := 0; iter < len(samples); iter++ {
			if len(verts) >= maxVerts {
				break
			}

			var bestPoint DetailedVertex
			var bestd float64
			besti := -1

			for i := range samples {
				sample := samples[i]
				if sample.added {
					continue
				}

				// the sample location is jittered to get rid of some bad triangulations
				// which are caused by symmetrical data from the grid structure
				px := float64(sample.X)*sampleDist + getJitterX(i)*cs*0.1
				py := float64(sample.Y) * ch
				pz := float64(sample.Z)*sampleDist + getJitterZ(i)*cs*0.1
				pt := DetailedVertex{X: px, Y: py, Z: pz}

				d := distToTris(pt, verts, tris)
				if d < 0 {
					continue
				}
				if d > bestd {
					bestd = d
					besti = i
					bestPoint = pt
				}
			}

			if bestd <= sampleMaxError || besti == -1 {
				break
			}

			samples[besti].added = true
			verts = append(verts, bestPoint)
			dmesh.InteriorSamples[polyIndex] = append(dmesh.InteriorSamples[polyIndex], float32(bestPoint.X), float32(bestPoint.Y), float32(bestPoint.Z))

			// TODO - incremental add instead of full rebuild
			tris = delaunayHull(verts, hull)
		}
	}

	setTriFlags(tris, hull)

	return verts, tris
}

func delaunayHull(verts []DetailedVertex, hull []int) []Triangle {
	var edges []DetailedEdge
	for i, j := 0, len(hull)-1; i < len(hull); j, i = i, i+1 {
		edges = addEdge(edges, hull[j], hull[i], edgeHull, edgeUndefined)
	}

	nfaces := 0

	e := 0
	for e < len(edges) {
		if edges[e].l == edgeUndefined {
			edges, nfaces = completeFacet(e, edges, verts, nfaces)
		}
		if edges[e].r == edgeUndefined {
			edges, nfaces = completeFacet(e, edges, verts, nfaces)
		}
		e++
	}

	tris := make([]Triangle, nfaces)
	for i := range len(tris) {
		t := &tris[i]
		t.Vertices = [3]int{-1, -1, -1}
	}

	for _, edge := range edges {
		if edge.r >= 0 {
			// left face
			t := &tris[edge.r]
			if t.Vertices[0] == -1 {
				t.Vertices[0] = edge.s
				t.Vertices[1] = edge.t
			} else if t.Vertices[0] == edge.t {
				t.Vertices[2] = edge.s
			} else if t.Vertices[1] == edge.s {
				t.Vertices[2] = edge.t
			}
		}
		if edge.l >= 0 {
			// right face
			t := &tris[edge.l]
			if t.Vertices[0] == -1 {
				t.Vertices[0] = edge.t
				t.Vertices[1] = edge.s
			} else if t.Vertices[0] == edge.s {
				t.Vertices[2] = edge.t
			} else if t.Vertices[1] == edge.t {
				t.Vertices[2] = edge.s
			}
		}
	}

	// remove dangling face
	for i := 0; i < len(tris); i++ {
		t := &tris[i]
		if t.Vertices[0] == -1 || t.Vertices[1] == -1 || t.Vertices[2] == -1 {
			t.Vertices[0] = tris[len(tris)-1].Vertices[0]
			t.Vertices[1] = tris[len(tris)-1].Vertices[1]
			t.Vertices[2] = tris[len(tris)-1].Vertices[2]
			t.OnHull = tris[len(tris)-1].OnHull
			fmt.Printf("removing dangling face %d [%d, %d, %d]\n", i, t.Vertices[0], t.Vertices[1], t.Vertices[2])

			tris = tris[:len(tris)-1]
			i--
		}
	}

	return tris
}

func setTriFlags(tris []Triangle, hull []int) {
	for i := range tris {
		tri := &tris[i]
		tri.OnHull[0] = onHull(tri.Vertices[0], tri.Vertices[1], hull)
		tri.OnHull[1] = onHull(tri.Vertices[1], tri.Vertices[2], hull)
		tri.OnHull[2] = onHull(tri.Vertices[2], tri.Vertices[0], hull)
	}
}

func onHull(a, b int, hull []int) bool {
	// all internal sampled points come after the hull so we can early out for those
	if a >= len(hull) || b >= len(hull) {
		return false
	}

	for i, j := 0, len(hull)-1; i < len(hull); j, i = i, i+1 {
		if a == hull[j] && b == hull[i] {
			return true
		}
	}

	return false
}

func completeFacet(e int, edges []DetailedEdge, verts []DetailedVertex, f int) ([]DetailedEdge, int) {
	var epsilon float64 = 1e-5
	var tolerance float64 = 1e-3

	edge := &edges[e]

	// cache s and t
	var s, t int

	if edge.l == edgeUndefined {
		s = edge.s
		t = edge.t
		// s = edge.t
		// t = edge.s
	} else if edge.r == edgeUndefined {
		s = edge.t
		t = edge.s
		// s = edge.s
		// t = edge.t
	} else {
		return edges, 0
	}

	// find the best point on the left of edge

	pt := len(verts)
	var center DetailedVertex
	var radius float64 = -1

	// iterate over all points that are not part of the edge
	// and find the point that produces the smallest circumcircle
	for u := range len(verts) {
		if u == s || u == t {
			continue
		}

		if vCross2D(verts[s], verts[t], verts[u]) > epsilon {
			// if vCross2D(verts[s], verts[t], verts[u]) < epsilon {
			if radius < 0 {
				// first time circumcircle is set
				pt = u
				center, radius = circumCircle(verts[s], verts[t], verts[u])
				continue
			}

			d := dist2D(center, verts[u])
			if d > radius*(1.0+tolerance) {
				// outside of the current circumcircle
				continue
			} else if d < radius*(1.0-tolerance) {
				pt = u
				center, radius = circumCircle(verts[s], verts[t], verts[u])
				continue
			} else {
				// if we're within some margin of error for the circumcircle
				// conduct some more tests
				if overlapEdges(verts, edges, s, u) {
					continue
				}
				if overlapEdges(verts, edges, t, u) {
					continue
				}
				// edge is valid
				pt = u
				center, radius = circumCircle(verts[s], verts[t], verts[u])
			}
		}
	}

	// add a new triangle or update the edge info if s-t is on the hull
	if pt < len(verts) {
		// update face information being completed
		updateLeftFace(&edges[e], s, t, f)

		// add a new edge or update the face info of the old edge
		e := findEdge(edges, pt, s)
		if e == edgeUndefined {
			edges = addEdge(edges, pt, s, f, edgeUndefined)
		} else {
			updateLeftFace(&edges[e], pt, s, f)
		}

		// add a new edge or update the face info of the old edge
		e = findEdge(edges, t, pt)
		if e == edgeUndefined {
			edges = addEdge(edges, t, pt, f, edgeUndefined)
		} else {
			updateLeftFace(&edges[e], t, pt, f)
		}
		f++
	} else {
		updateLeftFace(&edges[e], s, t, edgeHull)
	}

	return edges, f
}

func updateLeftFace(edge *DetailedEdge, s, t, f int) {
	if edge.s == s && edge.t == t && edge.l == edgeUndefined {
		edge.l = f
	} else if edge.t == s && edge.s == t && edge.r == edgeUndefined {
		edge.r = f
	}
}

func overlapSegSeg2D(a, b, c, d DetailedVertex) bool {
	a1 := vCross2D(a, b, d)
	a2 := vCross2D(a, b, c)
	if a1*a2 < 0 {
		a3 := vCross2D(c, d, a)
		a4 := a3 + a2 - a1
		if a3*a4 < 0 {
			return true
		}
	}
	return false
}

func overlapEdges(verts []DetailedVertex, edges []DetailedEdge, s1, t1 int) bool {
	for _, edge := range edges {
		s0 := edge.s
		t0 := edge.t

		// same or connected edges do not overlap
		if s0 == s1 || s0 == t1 || t0 == s1 || t0 == t1 {
			continue
		}
		if overlapSegSeg2D(verts[s0], verts[t0], verts[s1], verts[t1]) {
			return true
		}
	}
	return false
}

func circumCircle(p1, p2, p3 DetailedVertex) (DetailedVertex, float64) {
	var epsilon float64 = 1e-6

	// calculate the circle relative to avoid some precision issues
	var v1 DetailedVertex
	v2 := vSub(p2, p1)
	v3 := vSub(p3, p1)

	var center DetailedVertex
	var radius float64

	cp := vCross2D(v1, v2, v3)
	if math.Abs(cp) > epsilon {
		v1Sq := vDot2D(v1, v1)
		v2Sq := vDot2D(v2, v2)
		v3Sq := vDot2D(v3, v3)

		center.X = (v1Sq*(v2.Z-v3.Z) + v2Sq*(v3.Z-v1.Z) + v3Sq*(v1.Z-v2.Z)) / (2 * cp)
		center.Y = 0
		center.Z = (v1Sq*(v3.X-v2.X) + v2Sq*(v1.X-v3.X) + v3Sq*(v2.X-v1.X)) / (2 * cp)

		radius = dist2D(center, v1)

		center.X += p1.X
		center.Y += p1.Y
		center.Z += p1.Z
	}

	return center, radius
}

func addEdge(edges []DetailedEdge, s, t, l, r int) []DetailedEdge {
	e := findEdge(edges, s, t)
	if e == edgeUndefined {
		result := edges
		result = append(result, DetailedEdge{
			s: s,
			t: t,
			l: l,
			r: r,
		})
		return result
	}

	return edges
}

func findEdge(edges []DetailedEdge, s, t int) int {
	for i, edge := range edges {
		if edge.s == s && edge.t == t || edge.s == t && edge.t == s {
			return i
		}
	}
	return edgeUndefined
}

func vSub(a, b DetailedVertex) DetailedVertex {
	return DetailedVertex{X: a.X - b.X, Y: a.Y - b.Y, Z: a.Z - b.Z}
}

func vDot2D(a, b DetailedVertex) float64 {
	return a.X*b.X + a.Z*b.Z
}

func distToTri(p, a, b, c DetailedVertex) float64 {
	v0 := vSub(c, a)
	v1 := vSub(b, a)
	v2 := vSub(p, a)

	dot00 := vDot2D(v0, v0)
	dot01 := vDot2D(v0, v1)
	dot02 := vDot2D(v0, v2)
	dot11 := vDot2D(v1, v1)
	dot12 := vDot2D(v1, v2)

	// compute barycentric coordinates
	var invDenom float64 = 1.0 / (dot00*dot11 - dot01*dot01)
	var u float64 = (dot11*dot02 - dot01*dot12) * invDenom
	var v float64 = (dot00*dot12 - dot01*dot02) * invDenom

	// if point lies inside of the triangle, return the interpolated y -coord
	epsilon := 1e-4
	if u >= -epsilon && v >= -epsilon && (u+v) <= 1+epsilon {
		y := a.Y + v0.Y*u + v1.Y*v
		return math.Abs(y - p.Y)
	}

	return math.MaxFloat64
}

func distToTris(p DetailedVertex, verts []DetailedVertex, tris []Triangle) float64 {
	dmin := math.MaxFloat64
	for i := range len(tris) {
		a := verts[tris[i].Vertices[0]]
		b := verts[tris[i].Vertices[1]]
		c := verts[tris[i].Vertices[2]]
		d := distToTri(p, a, b, c)
		if d < dmin {
			dmin = d
		}
	}
	if dmin == math.MaxFloat64 {
		return -1
	}
	return dmin
}

func distToPoly(verts []DetailedVertex, p DetailedVertex) float64 {
	n := len(verts)
	c := false
	dmin := math.MaxFloat64
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		vi := verts[i]
		vj := verts[j]
		if ((vi.Z > p.Z) != (vj.Z > p.Z)) && (p.X < (vj.X-vi.X)*(p.Z-vi.Z)/(vj.Z-vi.Z)+vi.X) {
			c = !c
		}
		d, _ := distancePtSeg2Df(p.X, p.Z, vj.X, vj.Z, vi.X, vi.Z)
		dmin = min(dmin, d)
	}
	if c {
		return -dmin
	}
	return dmin
}

func dvMin(v0, v1 *DetailedVertex) {
	v0.X = min(v0.X, v1.X)
	v0.Y = min(v0.Y, v1.Y)
	v0.Z = min(v0.Z, v1.Z)
}

func dvMax(v0, v1 *DetailedVertex) {
	v0.X = max(v0.X, v1.X)
	v0.Y = max(v0.Y, v1.Y)
	v0.Z = max(v0.Z, v1.Z)
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
	tris = append(tris, Triangle{Vertices: [3]int{hull[start], hull[left], hull[right]}})

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
			tris = append(tris, Triangle{Vertices: [3]int{hull[left], hull[nLeft], hull[right]}})
			left = nLeft
		} else {
			tris = append(tris, Triangle{Vertices: [3]int{hull[left], hull[nRight], hull[right]}})
			right = nRight
		}
	}

	return tris
}

var DBG []int

func getHeight(fx, fy, fz, ics, ch float64, radius int, hp HeightPatch) int {
	ix := int(math.Floor(fx*ics + 0.01))
	iz := int(math.Floor(fz*ics + 0.01))

	ix = Clamp(ix-hp.xmin, 0, hp.width-1)
	iz = Clamp(iz-hp.zmin, 0, hp.height-1)

	h := hp.data[ix+iz*hp.width]

	// spiral search for next available height value
	if h == hpUnsetHeight {
		DBG = append(DBG, ix+hp.xmin, iz+hp.zmin)
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
		// panic("failed to find height")
		fmt.Printf("failed to sample %.0f %.0f %.0f\n", fx, fy, fz)
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
			d, _ := distancePtSeg2Df(verts[j].X, verts[j].Z, p1.X, p1.Z, p2.X, p2.Z)
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
func distancePtSeg2Df(x, z, px, pz, qx, qz float64) (float64, float64) {
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

	return dx*dx + dz*dz, t
}

func dist2D(v0, v1 DetailedVertex) float64 {
	return math.Sqrt(dist2dSq(v0, v1))
}

func dist2dSq(v0, v1 DetailedVertex) float64 {
	dx := v0.X - v1.X
	dz := v0.Z - v1.Z
	return dx*dx + dz*dz
}

func vCross2D(p1, p2, p3 DetailedVertex) float64 {
	u1 := p2.X - p1.X
	v1 := p2.Z - p1.Z
	u2 := p3.X - p1.X
	v2 := p3.Z - p1.Z
	return u1*v2 - v1*u2
}

func getJitterX(i int) float64 {
	return (float64((i*0x8da6b343)&0xffff) / 65535.0 * 2.0) - 1.0
}

func getJitterZ(i int) float64 {
	return (float64((i*0xd8163841)&0xffff) / 65535.0 * 2.0) - 1.0
}

func next(i, max int) int {
	return (i + 1) % max
}

func prev(i, max int) int {
	return (i - 1 + max) % max
}
