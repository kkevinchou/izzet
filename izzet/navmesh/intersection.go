package navmesh

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

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
		return false
	}

	// a01, p1 == p2
	p0 = -v0[1]*(v2[2]-v1[2]) + v0[2]*(v2[1]-v1[1])
	p1 = -v1[1]*v2[2] + v1[2]*v2[1]
	r = extents[1]*math.Abs(f1[2]) + extents[2]*math.Abs(f1[1])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		return false
	}

	// a02, p0 == p2
	p0 = v0[1]*v2[2] - v0[2]*v2[1]
	p1 = -v1[1]*(v0[2]-v2[2]) + v1[2]*(v0[1]-v2[1])
	r = extents[1]*math.Abs(f2[2]) + extents[2]*math.Abs(f2[1])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		return false
	}

	//////////////////////////////////////////////////

	// a10, p0 == p1
	p0 = v0[0]*v1[2] - v0[2]*v1[0]
	p2 = v2[0]*(v1[2]-v0[2]) - v2[2]*(v1[0]-v0[0])
	r = extents[0]*math.Abs(f0[2]) + extents[2]*math.Abs(f0[0])
	if math.Max(-math.Max(p0, p2), math.Min(p0, p2)) > r {
		return false
	}

	// a11, p1 == p2

	p0 = v0[0]*(v2[2]-v1[2]) - v0[2]*(v2[0]-v1[0])
	p1 = v1[0]*v2[2] - v1[2]*v2[0]
	r = extents[0]*math.Abs(f1[2]) + extents[2]*math.Abs(f1[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		return false
	}

	// a12, p0 == p2

	p0 = -v0[0]*v2[2] + v0[2]*v2[0]
	p1 = v1[0]*(v0[2]-v2[2]) - v1[2]*(v0[0]-v2[0])
	r = extents[0]*math.Abs(f2[2]) + extents[2]*math.Abs(f2[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		return false
	}

	//////////////////////////////////////////////////

	// a20, p0 == p1

	p0 = -v0[0]*v1[1] + v0[1]*v1[0]
	p2 = -v2[0]*(v1[1]-v0[1]) + v2[1]*(v1[0]-v0[0])
	r = extents[0]*math.Abs(f0[1]) + extents[1]*math.Abs(f0[0])
	if math.Max(-math.Max(p0, p2), math.Min(p0, p2)) > r {
		return false
	}

	p0 = -v0[0]*(v2[1]-v1[1]) + v0[1]*(v2[0]-v1[0])
	p1 = -v1[0]*v2[1] + v1[1]*v2[0]
	r = extents[0]*math.Abs(f1[1]) + extents[1]*math.Abs(f1[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		return false
	}

	p0 = v0[0]*v2[1] - v0[1]*v2[0]
	p1 = -v1[0]*(v0[1]*v2[1]) + v1[1]*(v0[0]-v2[0])
	r = extents[0]*math.Abs(f2[1]) + extents[1]*math.Abs(f2[0])
	if math.Max(-math.Max(p0, p1), math.Min(p0, p1)) > r {
		return false
	}

	//////////////////////////////////////////////////

	// test against each AABB face normal

	if math.Max(v0[0], math.Max(v1[0], v2[0])) < -extents[0] || math.Min(v0[0], math.Min(v1[0], v2[0])) > extents[0] {
		return false
	}
	if math.Max(v0[1], math.Max(v1[1], v2[1])) < -extents[1] || math.Min(v0[1], math.Min(v1[1], v2[1])) > extents[1] {
		return false
	}
	if math.Max(v0[2], math.Max(v1[2], v2[2])) < -extents[2] || math.Min(v0[2], math.Min(v1[2], v2[2])) > extents[2] {
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
