package navmesh

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

const (
	vertexBucketCount = (1 << 12)
	nvp               = 6 // max verts per poly
)

type Index struct {
	index     int
	removable bool
}

type PolyTriangle struct {
	a int
	b int
	c int
}

type PolyVertex struct {
	X, Y, Z int
}

type Polygon struct {
	// index to the vertices owned by Mesh that make up this polygon
	Verts []int
	// polyNeighbor[i] stores the polygon index sharing the edge (i, i+1), defined by
	// the vertices i and i+1
	polyNeighbor []int

	RegionID int
}

type Mesh struct {
	Vertices          []PolyVertex
	Polygons          []Polygon
	PremergeTriangles []Polygon
	areas             []AREA_TYPE
	maxEdgeError      float64
	bMin, bMax        mgl64.Vec3
}

func BuildPolyMesh(contourSet *ContourSet) *Mesh {
	mesh := &Mesh{
		maxEdgeError: contourSet.maxError,
		bMin:         contourSet.bMin,
		bMax:         contourSet.bMax,
	}

	maxVertices := 0
	maxTris := 0
	maxVertsPerContour := 0

	for _, contour := range contourSet.Contours {
		if len(contour.Verts) < 3 {
			continue
		}
		maxVertices += len(contour.Verts)
		maxTris += len(contour.Verts) - 2
		maxVertsPerContour = max(maxVertsPerContour, len(contour.Verts))
	}

	firstVert := make([]int, vertexBucketCount)
	for i := range firstVert {
		firstVert[i] = -1
	}
	nextVert := make([]int, maxVertices)

	for _, contour := range contourSet.Contours {
		if len(contour.Verts) < 3 {
			continue
		}

		tris := triangulate(contour.Verts)
		if len(tris) <= 0 {
			fmt.Println("bad triangulation")
		}

		var indices []int
		for j := 0; j < len(contour.Verts); j++ {
			v := contour.Verts[j]
			indices = append(indices, mesh.addVertex(v.X, v.Y, v.Z, firstVert, nextVert))
		}

		var polygons []Polygon

		for j := 0; j < len(tris); j++ {
			t := tris[j]
			if t.a != t.b && t.a != t.c && t.b != t.c {
				p := Polygon{
					Verts:        []int{indices[t.a], indices[t.b], indices[t.c]},
					polyNeighbor: []int{-1, -1, -1},
					RegionID:     contour.RegionID,
				}
				polygons = append(polygons, p)
				mesh.PremergeTriangles = append(mesh.PremergeTriangles, p)
			}
		}

		if len(polygons) == 0 {
			continue
		}

		// merge polygons
		for {
			bestMergeVal := 0
			bestPa, bestPb, bestEa, bestEb := 0, 0, 0, 0
			for j := range len(polygons) - 1 {
				pj := polygons[j]
				for k := j + 1; k < len(polygons); k++ {
					pk := polygons[k]
					v, ea, eb := getPolyMergeValue(pj, pk, mesh.Vertices)
					if v > bestMergeVal {
						bestMergeVal = v
						bestPa = j
						bestPb = k
						bestEa = ea
						bestEb = eb
					}
				}
			}

			// found the best merge candidates
			if bestMergeVal > 0 {
				pa := polygons[bestPa]
				pb := polygons[bestPb]

				// merge verts
				polygons[bestPa].Verts = mergePolyVerts(pa, pb, bestEa, bestEb)
				polygons[bestPa].polyNeighbor = make([]int, len(polygons[bestPa].Verts))

				// swap the last polygon to where b used to be, and reduce the polygon list size
				if bestPb != len(polygons)-1 {
					polygons[bestPb] = polygons[len(polygons)-1]
				}
				polygons = polygons[:len(polygons)-1]
			} else {
				break
			}
		}

		// store polygons
		for _, polygon := range polygons {
			mesh.areas = append(mesh.areas, contour.area)
			mesh.Polygons = append(mesh.Polygons, polygon)
			if len(mesh.Polygons) > maxTris {
				panic(fmt.Sprintf("too many polygons %d, max: %d", len(mesh.Polygons), maxTris))
			}
		}
	}

	// TODO - remove edge vertices for border vertexes
	// done by checking if the vertex has the `borderVertexFlag` flag set

	// calculate adjacency
	buildMeshAdjacency(mesh.Polygons, len(mesh.Vertices))

	// TODO - find portal edges. only runs when on borderSize > 0

	return mesh
}

func mergePolyVerts(pa, pb Polygon, ea, eb int) []int {
	na := len(pa.Verts)
	nb := len(pb.Verts)
	var mergedVerts []int

	for i := range len(pa.Verts) - 1 {
		mergedVerts = append(mergedVerts, pa.Verts[(ea+1+i)%na])
	}
	for i := range len(pb.Verts) - 1 {
		mergedVerts = append(mergedVerts, pb.Verts[(eb+1+i)%nb])
	}

	return mergedVerts
}

// merge with a polygon with the greatest edge length
func getPolyMergeValue(pa, pb Polygon, verts []PolyVertex) (int, int, int) {
	na := len(pa.Verts)
	nb := len(pb.Verts)

	// sum of the vertices - 2. 2 shared vertices will be removed
	if na+nb-2 > nvp {
		return -1, -1, -1
	}

	ea, eb := -1, -1

	found := false
	for i := 0; i < na; i++ {
		va0 := pa.Verts[i]
		va1 := pa.Verts[(i+1)%na]

		if va0 > va1 {
			va0, va1 = va1, va0
		}

		for j := 0; j < nb; j++ {
			vb0 := pb.Verts[j]
			vb1 := pb.Verts[(j+1)%nb]
			if vb0 > vb1 {
				vb0, vb1 = vb1, vb0
			}

			// found a shared edge
			if va0 == vb0 && va1 == vb1 {
				ea = i
				eb = j
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if ea == -1 || eb == -1 {
		return -1, -1, -1
	}

	// check if the merged polygon would be convex
	va := pa.Verts[(ea+na-1)%na]
	vb := pa.Verts[ea]
	vc := pb.Verts[(eb+2)%nb]
	if !uLeft(verts[va], verts[vb], verts[vc]) {
		return -1, -1, -1
	}

	va = pb.Verts[(eb+nb-1)%nb]
	vb = pb.Verts[eb]
	vc = pa.Verts[(ea+2)%na]
	if !uLeft(verts[va], verts[vb], verts[vc]) {
		return -1, -1, -1
	}

	va = pa.Verts[ea]
	vb = pa.Verts[(ea+1)%na]

	dx := verts[va].X - verts[vb].X
	dz := verts[va].Z - verts[vb].Z

	return dx*dx + dz*dz, ea, eb
}

func uLeft(a, b, c PolyVertex) bool {
	return (b.X-a.X)*(c.Z-a.Z)-(c.X-a.X)*(b.Z-a.Z) < 0
}

type Edge struct {
	vert     [2]int
	poly     [2]int
	polyEdge [2]int
}

func buildMeshAdjacency(polygons []Polygon, numVerts int) {
	maxEdgeCount := len(polygons) * nvp

	firstEdge := make([]int, numVerts)
	nextEdge := make([]int, maxEdgeCount)

	for i := range firstEdge {
		firstEdge[i] = -1
	}
	for i := range nextEdge {
		nextEdge[i] = -1
	}

	var edgeCount int
	var edges []Edge

	for i, polygon := range polygons {
		for j := 0; j < len(polygon.Verts); j++ {
			v0 := polygon.Verts[j]
			v1 := polygon.Verts[(j+1)%len(polygon.Verts)]
			if v0 < v1 {
				edge := Edge{
					vert:     [2]int{v0, v1},
					poly:     [2]int{i, i},
					polyEdge: [2]int{j, 0},
				}
				edges = append(edges, edge)

				nextEdge[edgeCount] = firstEdge[v0]
				firstEdge[v0] = edgeCount
				edgeCount++
			}
		}
	}

	for i, polygon := range polygons {
		for j := 0; j < len(polygon.Verts); j++ {
			v0 := polygon.Verts[j]
			v1 := polygon.Verts[(j+1)%len(polygon.Verts)]
			if v0 > v1 {
				for e := firstEdge[v1]; e != -1; e = nextEdge[e] {
					edge := &edges[e]
					if edge.vert[1] == v0 && edge.poly[0] == edge.poly[1] {
						edge.poly[1] = i
						edge.polyEdge[1] = j
						break
					}
				}
			}
		}
	}

	for i := range edgeCount {
		edge := &edges[i]
		if edge.poly[0] != edge.poly[1] {
			p0 := &polygons[edge.poly[0]]
			p1 := &polygons[edge.poly[1]]

			p0.polyNeighbor[edge.polyEdge[0]] = edge.poly[1]
			p1.polyNeighbor[edge.polyEdge[1]] = edge.poly[0]
		}
	}
}

func (m *Mesh) addVertex(x, y, z int, firstVert, nextVert []int) int {
	bucket := computeVertexHash(x, y, z)
	i := firstVert[bucket]

	for i != -1 {
		v := m.Vertices[i]
		if v.X == x && (math.Abs(float64(v.Y-y)) <= 2) && v.Z == z {
			return i
		}
		i = nextVert[i]
	}

	i = len(m.Vertices)
	m.Vertices = append(m.Vertices, PolyVertex{X: x, Y: y, Z: z})
	nextVert[i] = firstVert[bucket]
	firstVert[bucket] = i

	return i
}

func triangulate(vertices []SimplifiedVertex) []PolyTriangle {
	indices := make([]Index, len(vertices))
	for i := range indices {
		indices[i].index = i
	}

	for i := range len(vertices) {
		i1 := next(i, len(vertices))
		i2 := next(i1, len(vertices))

		if diagonal(i, i2, len(vertices), vertices, indices) {
			indices[i1].removable = true
		}
	}

	var tris []PolyTriangle
	n := len(vertices)
	for n > 3 {
		minLength := -1
		mini := -1

		for i := 0; i < n; i++ {
			i1 := next(i, n)
			if indices[i1].removable {
				p0 := vertices[indices[i].index]
				p2 := vertices[indices[next(i1, n)].index]

				dx := p2.X - p0.X
				dz := p2.Z - p0.Z
				len := dx*dx + dz*dz

				if minLength < 0 || len < minLength {
					minLength = len
					mini = i
				}
			}
		}

		if mini == -1 {
			minLength = -1
			mini = -1
			for i := 0; i < n; i++ {
				i1 := next(i, n)
				i2 := next(i1, n)
				if diagonalLoose(i, i2, n, vertices, indices) {
					p0 := vertices[indices[i].index]
					p2 := vertices[indices[next(i2, n)].index]
					dx := p2.X - p0.X
					dy := p2.Z - p0.Z
					len := dx*dx + dy*dy

					if minLength < 0 || len < minLength {
						minLength = len
						mini = i
					}
				}
			}
		}

		if mini == -1 {
			fmt.Println("failed to triangulate")
			return nil
		}

		i := mini
		i1 := next(i, n)
		i2 := next(i1, n)

		tris = append(tris, PolyTriangle{a: indices[i].index, b: indices[i1].index, c: indices[i2].index})

		n--
		for j := i1; j < n; j++ {
			indices[j] = indices[j+1]
		}

		if i1 >= n {
			i1 = 0
		}

		i = prev(i1, n)
		indices[i].removable = diagonal(prev(i, n), i1, n, vertices, indices)
		indices[i1].removable = diagonal(i, next(i1, n), n, vertices, indices)
	}

	tris = append(tris, PolyTriangle{a: indices[0].index, b: indices[1].index, c: indices[2].index})

	return tris
}

func diagonalLoose(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	return isConeLoose(i, j, n, vertices, indices) && diagonalieLoose(i, j, n, vertices, indices)
}

func isConeLoose(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	pi := vertices[indices[i].index]
	pj := vertices[indices[j].index]

	previ := prev(i, n)
	nexti := next(i, n)

	pprev := vertices[indices[previ].index]
	pnext := vertices[indices[nexti].index]

	if leftOn(pprev, pi, pnext) {
		return leftOn(pi, pj, pprev) && leftOn(pj, pi, pnext)
	}

	l1 := leftOn(pi, pj, pnext)
	l2 := leftOn(pj, pi, pprev)

	return !(l1 && l2)
}

func diagonalieLoose(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	d0 := vertices[indices[i].index]
	d1 := vertices[indices[j].index]

	for k := range n {
		k1 := (k + 1) % n

		if k == i || k == j || k1 == i || k1 == j {
			continue
		}

		p0 := vertices[indices[k].index]
		p1 := vertices[indices[k1].index]

		if vequal(d0, p0) || vequal(d1, p0) || vequal(d0, p1) || vequal(d1, p1) {
			continue
		}

		if intersectProp(d0, d1, p0, p1) {
			return false
		}
	}

	return true
}

// returns true if the line segment i-j is a proper internal diagonal
func diagonal(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	c := inCone(i, j, n, vertices, indices)
	d := diagonalie(i, j, n, vertices, indices)
	return c && d
}

// returns true iff the diagonal (i,j) is strictly internal to the
// polygon P in the neighborhood of the i endpoint.
func inCone(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	pi := vertices[indices[i].index]
	pj := vertices[indices[j].index]

	previ := prev(i, n)
	nexti := next(i, n)

	pprev := vertices[indices[previ].index]
	pnext := vertices[indices[nexti].index]

	if leftOn(pprev, pi, pnext) {
		return left(pi, pj, pprev) && left(pj, pi, pnext)
	}

	l1 := leftOn(pi, pj, pnext)
	l2 := leftOn(pj, pi, pprev)

	return !(l1 && l2)
}

// returns T iff (v_i, v_j) is a proper internal *or* external
// diagonal of P, *ignoring edges incident to v_i and v_j*.
func diagonalie(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	d0 := vertices[indices[i].index]
	d1 := vertices[indices[j].index]

	for k := range n {
		k1 := (k + 1) % n

		if k == i || k == j || k1 == i || k1 == j {
			continue
		}

		p0 := vertices[indices[k].index]
		p1 := vertices[indices[k1].index]

		if vequal(d0, p0) || vequal(d1, p0) || vequal(d0, p1) || vequal(d1, p1) {
			continue
		}

		if intersect(d0, d1, p0, p1) {
			return false
		}
	}

	return true
}

func between(a, b, c SimplifiedVertex) bool {
	if !colinear(a, b, c) {
		return false
	}

	if a.X != b.X {
		return (a.X <= c.X && c.X <= b.X) || (a.X >= c.X && c.X >= b.X)
	}

	return (a.Z <= c.Z && c.Z <= b.Z) || (a.Z >= c.Z && c.Z >= b.Z)

}

// returns true iff segments ab and cd intersect, properly or improperly.
func intersect(a, b, c, d SimplifiedVertex) bool {
	if intersectProp(a, b, c, d) {
		return true
	}

	if between(a, b, c) || between(a, b, d) || between(c, d, a) || between(c, d, b) {
		return true
	}

	return false
}

// returns true iff ab properly intersects cd: they share
// a point interior to both segments.  The properness of the
// intersection is ensured by using strict leftness.
func intersectProp(a, b, c, d SimplifiedVertex) bool {
	if colinear(a, b, c) || colinear(a, b, d) || colinear(c, d, a) || colinear(c, d, b) {
		return false
	}

	return xorb(left(a, b, c), left(a, b, d)) && xorb(left(c, d, a), left(c, d, b))
}

func xorb(a, b bool) bool {
	return (a || b) && !(a && b)
}

func vequal(a, b SimplifiedVertex) bool {
	return a.X == b.X && a.Z == b.Z
}

func leftOn(a, b, c SimplifiedVertex) bool {
	return area2D(a, b, c) <= 0
}

func left(a, b, c SimplifiedVertex) bool {
	return area2D(a, b, c) < 0
}

func colinear(a, b, c SimplifiedVertex) bool {
	return area2D(a, b, c) == 0
}

func area2D(a, b, c SimplifiedVertex) int {
	p := (b.X - a.X) * (c.Z - a.Z)
	q := (c.X - a.X) * (b.Z - a.Z)
	value := p - q
	return value
}

func computeVertexHash(x, y, z int) int {
	h1 := 0x8da6b343 // Large multiplicative constants;
	h2 := 0xd8163841 // here arbitrarily chosen primes
	h3 := 0xcb1ab31f
	n := h1*x + h2*y + h3*z

	// ensure hash is always positive
	return (((n % vertexBucketCount) + vertexBucketCount) % vertexBucketCount)
}
