package navmesh

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

func triMax(a, b, c float64) float64 {
	return math.Max(a, math.Max(b, c))
}
func triMin(a, b, c float64) float64 {
	return math.Min(a, math.Min(b, c))
}

// IntersectAABBTriangle implements the algorithm from Real Time Collision Detection - "Testing AABB Against Triangle"
func IntersectAABBTriangle(aabb AABB, tri Triangle) bool {
	// Translate the AABB is centered at the origin.
	center := aabb.Min.Add(aabb.Max).Mul(0.5)
	extents := aabb.Max.Sub(center)

	// Translate the triangle to be relative to the origin
	v0 := tri.V1.Sub(center)
	v1 := tri.V2.Sub(center)
	v2 := tri.V3.Sub(center)

	// Compute edge vectors for triangle.
	f0 := v1.Sub(v0)
	f1 := v2.Sub(v1)
	f2 := v0.Sub(v2)

	var p0, p1, p2, r float64

	// Test against the nine axes given by the cross products between the AABB normals and triangle edges

	// a00, p0 == p1
	p0 = v0[2]*v1[1] - v0[1]*v1[2]
	p2 = v2[2]*(v1[1]-v0[1]) - v2[2]*(v1[2]-v0[2])
	r = extents[1]*math.Abs(f0[2]) + extents[2]*math.Abs(f0[1])
	if math.Max(-math.Max(p0, p2), math.Min(p0, p2)) > r {
		a00 := mgl64.Vec3{0, -f0[2], f0[1]}
		p0 = v0.Dot(a00)
		p1 = v1.Dot(a00)
		p2 = v2.Dot(a00)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a00")
		}
		return false
	}

	// a01, p1 == p2
	p0 = -v0[1]*(v2[2]-v1[2]) + v0[2]*(v2[1]-v1[1])
	p1 = -v1[1]*v2[2] + v1[2]*v2[1]
	r = extents[1]*math.Abs(f1[2]) + extents[2]*math.Abs(f1[1])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		a01 := mgl64.Vec3{0, -f1[2], f1[1]}
		p0 = v0.Dot(a01)
		p1 = v1.Dot(a01)
		p2 = v2.Dot(a01)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a01")
		}
		return false
	}

	// a02, p0 == p2
	p0 = v0[1]*v2[2] - v0[2]*v2[1]
	p1 = -v1[1]*(v0[2]-v2[2]) + v1[2]*(v0[1]-v2[1])
	r = extents[1]*math.Abs(f2[2]) + extents[2]*math.Abs(f2[1])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		a02 := mgl64.Vec3{0, -f2[2], f2[1]}
		p0 = v0.Dot(a02)
		p1 = v1.Dot(a02)
		p2 = v2.Dot(a02)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a02")
		}
		return false
	}

	//////////////////////////////////////////////////

	// a10, p0 == p1
	p0 = v0[0]*v1[2] - v0[2]*v1[0]
	p2 = v2[0]*(v1[2]-v0[2]) - v2[2]*(v1[0]-v0[0])
	r = extents[0]*math.Abs(f0[2]) + extents[2]*math.Abs(f0[0])
	if math.Max(-math.Max(p0, p2), math.Min(p0, p2)) > r {
		a10 := mgl64.Vec3{f0[2], 0, -f0[0]}
		p0 = v0.Dot(a10)
		p1 = v1.Dot(a10)
		p2 = v2.Dot(a10)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a10")
		}
		return false
	}

	// a11, p1 == p2

	p0 = v0[0]*(v2[2]-v1[2]) - v0[2]*(v2[0]-v1[0])
	p1 = v1[0]*v2[2] - v1[2]*v2[0]
	r = extents[0]*math.Abs(f1[2]) + extents[2]*math.Abs(f1[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		a11 := mgl64.Vec3{f1[2], 0, -f1[0]}
		p0 = v0.Dot(a11)
		p1 = v1.Dot(a11)
		p2 = v2.Dot(a11)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a11")
		}
		return false
	}

	// a12, p0 == p2

	p0 = -v0[0]*v2[2] + v0[2]*v2[0]
	p1 = v1[0]*(v0[2]-v2[2]) - v1[2]*(v0[0]-v2[0])
	r = extents[0]*math.Abs(f2[2]) + extents[2]*math.Abs(f2[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		a12 := mgl64.Vec3{f2[2], 0, -f2[0]}
		p0 = v0.Dot(a12)
		p1 = v1.Dot(a12)
		p2 = v2.Dot(a12)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a12")
		}
		return false
	}

	//////////////////////////////////////////////////

	// a20, p0 == p1

	p0 = -v0[0]*v1[1] + v0[1]*v1[0]
	p2 = -v2[0]*(v1[1]-v0[1]) + v2[1]*(v1[0]-v0[0])
	r = extents[0]*math.Abs(f0[1]) + extents[1]*math.Abs(f0[0])
	if math.Max(-math.Max(p0, p2), math.Min(p0, p2)) > r {
		a20 := mgl64.Vec3{-f0[1], f0[0], 0}
		p0 = v0.Dot(a20)
		p1 = v1.Dot(a20)
		p2 = v2.Dot(a20)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a20")
		}
		return false
	}

	p0 = -v0[0]*(v2[1]-v1[1]) + v0[1]*(v2[0]-v1[0])
	p1 = -v1[0]*v2[1] + v1[1]*v2[0]
	r = extents[0]*math.Abs(f1[1]) + extents[1]*math.Abs(f1[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		a12 := mgl64.Vec3{-f1[1], f1[0], 0}
		p0 = v0.Dot(a12)
		p1 = v1.Dot(a12)
		p2 = v2.Dot(a12)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a21")
		}
		return false
	}

	p0 = v0[0]*v2[1] - v0[1]*v2[0]
	p1 = -v1[0]*(v0[1]-v2[1]) + v1[1]*(v0[0]-v2[0])
	r = extents[0]*math.Abs(f2[1]) + extents[1]*math.Abs(f2[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		a22 := mgl64.Vec3{-f2[1], f2[0], 0}
		p0 = v0.Dot(a22)
		p1 = v1.Dot(a22)
		p2 = v2.Dot(a22)
		check := math.Max(-triMax(p0, p1, p2), triMin(p0, p1, p2)) > r
		if !check {
			panic("a22")
		}
		return false
	}

	//////////////////////////////////////////////////

	// test against each AABB face normal

	if triMax(v0[0], v1[0], v2[0]) < -extents[0] || triMin(v0[0], v1[0], v2[0]) > extents[0] {
		return false
	}
	if triMax(v0[1], v1[1], v2[1]) < -extents[1] || triMin(v0[1], v1[1], v2[1]) > extents[1] {
		return false
	}
	if triMax(v0[2], v1[2], v2[2]) < -extents[2] || triMin(v0[2], v1[2], v2[2]) > extents[2] {
		return false
	}

	// test separating axis corresponding to triangle normal

	var plane Plane
	plane.n = f0.Cross(f1)
	plane.d = plane.n.Dot(v0)
	return TestAABBPlane(aabb, plane)

}
func TestAABBPlane(aabb AABB, plane Plane) bool {
	// Translate the AABB is centered at the origin.
	center := aabb.Min.Add(aabb.Max).Mul(0.5)
	extents := aabb.Max.Sub(center)

	var r float64 = extents[0]*math.Abs(plane.n[0]) + extents[1]*math.Abs(plane.n[1]) + extents[2]*math.Abs(plane.n[2])
	var s float64 = plane.n.Dot(center) - plane.d

	return math.Abs(s) <= r
}

type Plane struct {
	n mgl64.Vec3
	d float64
}

type AABB struct {
	Min, Max mgl64.Vec3
}

type Triangle struct {
	V1, V2, V3 mgl64.Vec3
}
