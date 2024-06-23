package navmesh

import "fmt"

type Index struct {
	index     int
	removable bool
}

type RCTriangle struct {
	a int
	b int
	c int
}

func BuildPolyMesh(contourSet *ContourSet) {
	for _, contour := range contourSet.Contours {
		if len(contour.Verts) < 3 {
			continue
		}

		triangulate(contour.Verts)
	}
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
