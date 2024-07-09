package navmesh

import (
	"fmt"
	"math"
)

const (
	vertexBucketCount = (1 << 12)
	nvp               = 1000 // max verts per poly
)

type Index struct {
	index     int
	removable bool
}

type RCTriangle struct {
	a int
	b int
	c int
}

type MeshVertex struct {
	x, y, z int
}

type PolyVertex struct {
	X, Y, Z int
}

type Polygon struct {
	// index to the vertices owned by Mesh that make up this polygon
	verts []int
	// polyNeighbor[i] stores the polygon index sharing the edge (i, i+1), defined by
	// the vertices i and i+1
	polyNeighbor []int
}

type Mesh struct {
	vertices  []PolyVertex
	polygons  []Polygon
	regionIDs []int
	areas     []AREA_TYPE
}

func BuildPolyMesh(contourSet *ContourSet) *Mesh {
	mesh := &Mesh{}

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
				polygons = append(polygons, Polygon{
					verts:        []int{indices[t.a], indices[t.b], indices[t.c]},
					polyNeighbor: []int{-1, -1, -1},
				})
			}
		}

		if len(polygons) == 0 {
			continue
		}

		// TODO - merge polygons

		// store polygons

		for _, polygon := range polygons {
			mesh.regionIDs = append(mesh.regionIDs, contour.RegionID)
			mesh.areas = append(mesh.areas, contour.area)
			mesh.polygons = append(mesh.polygons, polygon)
			if len(mesh.polygons) > maxTris {
				panic(fmt.Sprintf("too many polygons %d, max: %d", len(mesh.polygons), maxTris))
			}
		}
	}

	// TODO - remove edge vertices

	// calculate adjacency
	buildMeshAdjacency(mesh.polygons, len(mesh.vertices))

	// TODO - find portal edges

	return mesh
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
		for j := 0; j < len(polygon.verts); j++ {
			v0 := polygon.verts[j]
			v1 := polygon.verts[(j+1)%len(polygon.verts)]
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
		for j := 0; j < len(polygon.verts); j++ {
			v0 := polygon.verts[j]
			v1 := polygon.verts[(j+1)%len(polygon.verts)]
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
		v := m.vertices[i]
		if v.X == x && (math.Abs(float64(v.Y-y)) <= 2) && v.Z == z {
			return i
		}
		i = nextVert[i]
	}

	i = len(m.vertices)
	m.vertices = append(m.vertices, PolyVertex{X: x, Y: y, Z: z})
	nextVert[i] = firstVert[bucket]
	firstVert[bucket] = i

	return i
}

func triangulate(vertices []SimplifiedVertex) []RCTriangle {
	indices := make([]Index, len(vertices))
	for i := range indices {
		indices[i].index = i
	}

	for i := range len(vertices) {
		i1 := (i + 1) % len(vertices)
		i2 := (i + 2) % len(vertices)

		if diagonal(i, i2, len(vertices), vertices, indices) {
			indices[i1].removable = true
		}
	}

	var tris []RCTriangle
	vertCount := len(vertices)
	for vertCount > 3 {
		minLength := -1
		mini := -1

		for i := 0; i < vertCount; i++ {
			i1 := (i + 1) % vertCount
			i2 := (i + 2) % vertCount
			if indices[i1].removable {
				p0 := vertices[indices[i].index]
				p2 := vertices[indices[i2].index]

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
			fmt.Println("failed to triangulate")
			return nil
			// panic("WAT")
		}

		i := mini
		i1 := (i + 1) % vertCount
		i2 := (i + 2) % vertCount

		tris = append(tris, RCTriangle{a: indices[i].index, b: indices[i1].index, c: indices[i2].index})

		vertCount--
		for j := i1; j < vertCount; j++ {
			indices[j] = indices[j+1]
		}

		if i1 >= vertCount {
			i1 = 0
		}

		i = (i1 - 1 + vertCount) % vertCount
		previ := (i - 1 + vertCount) % vertCount
		i2 = (i1 + 1) % vertCount

		indices[i].removable = diagonal(previ, i1, vertCount, vertices, indices)
		indices[i1].removable = diagonal(i, i2, vertCount, vertices, indices)
	}

	tris = append(tris, RCTriangle{a: indices[0].index, b: indices[1].index, c: indices[2].index})

	return tris
}

// returns true if the line segment i-j is a proper internal diagonal
func diagonal(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	return inCone(i, j, n, vertices, indices) && diagonalie(i, j, n, vertices, indices)
}

func inCone(i, j, n int, vertices []SimplifiedVertex, indices []Index) bool {
	pi := vertices[indices[i].index]
	pj := vertices[indices[j].index]

	previ := (i - 1 + n) % n
	nexti := (i + 1) % n

	pprev := vertices[indices[previ].index]
	pnext := vertices[indices[nexti].index]

	if leftOn(pprev, pi, pnext) {
		return left(pi, pj, pprev) && left(pj, pi, pnext)
	}

	l1 := leftOn(pi, pj, pnext)
	l2 := leftOn(pj, pi, pprev)

	return !(l1 && l2)
}

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

func intersect(a, b, c, d SimplifiedVertex) bool {
	if intersectProp(a, b, c, d) {
		return true
	}

	if between(a, b, c) || between(a, b, d) || between(c, d, a) || between(c, d, b) {
		return true
	}

	return false
}

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
	return area2(a, b, c) <= 0
}

func left(a, b, c SimplifiedVertex) bool {
	return area2(a, b, c) < 0
}

func colinear(a, b, c SimplifiedVertex) bool {
	return area2(a, b, c) == 0
}

func area2(a, b, c SimplifiedVertex) int {
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
