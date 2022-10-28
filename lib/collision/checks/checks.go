package checks

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/lib/collision/collider"
)

const (
	epsilon float64 = 0.000001
)

func IntersectRayPlane(ray collider.Ray, plane collider.Plane) (*mgl64.Vec3, bool) {
	// front determines whether the plane is in front or behind the ray direction
	front := true

	directionDotNormal := ray.Direction.Dot(plane.Normal)
	if directionDotNormal == 0 {
		return nil, true
	}
	t := (plane.Point.Sub(ray.Origin)).Dot(plane.Normal)
	t /= directionDotNormal

	if t < 0 {
		front = false
	}

	intersectionPoint := ray.Origin.Add(ray.Direction.Mul(t))
	return &intersectionPoint, front
}

func IntersectRayTriangle(ray collider.Ray, triangle collider.Triangle) *mgl64.Vec3 {
	plane := collider.Plane{
		Point:  triangle.Points[0],
		Normal: triangle.Normal,
	}

	point, front := IntersectRayPlane(ray, plane)
	if point == nil || !front {
		return nil
	}

	if PointInTriangle(*point, triangle) {
		return point
	}

	return nil
}

func IntersectRayTriMesh(ray collider.Ray, triMesh collider.TriMesh) *mgl64.Vec3 {
	var minDist *float64
	var minPoint *mgl64.Vec3

	for _, t := range triMesh.Triangles {
		point := IntersectRayTriangle(ray, t)
		if point == nil {
			continue
		}

		if minDist == nil {
			dst := ray.Origin.Sub(*point).Len()
			minDist = &dst
			minPoint = point
		} else {
			dst := ray.Origin.Sub(*point).Len()
			if dst < *minDist {
				minDist = &dst
				minPoint = point
			}
		}
	}

	return minPoint
}

// ClosestPointOnLineToPoint returns the point on line segment AB that is closest
// to point C
func ClosestPointOnLineToPoint(a, b, c mgl64.Vec3) mgl64.Vec3 {
	ac := c.Sub(a)
	ab := b.Sub(a)

	t := ac.Dot(ab) / ab.Dot(ab)
	t = mgl64.Clamp(t, 0, 1)

	return a.Add(ab.Mul(t))
}

// 6 cases
// segment PQ and triangle edge AB,
// segment PQ and triangle edge BC,
// segment PQ and triangle edge CA,
// segment endpoint P and plane of triangle (when P projects inside ABC), and
// segment endpoint Q and plane of triangle (when Q projects inside ABC)

// the first point belongs to the triangle, the second point belongs to the line
func ClosestPointsLineVSTriangle(line collider.Line, triangle collider.Triangle) ([]mgl64.Vec3, float64) {
	var closestPoints []mgl64.Vec3
	var closestDistance float64

	a := triangle.Points[0]
	b := triangle.Points[1]
	c := triangle.Points[2]

	abPoints, abDist := ClosestPointsLineVSLine(line, collider.Line{P1: a, P2: b})
	closestPoints = abPoints
	closestDistance = abDist

	bcPoints, bcDist := ClosestPointsLineVSLine(line, collider.Line{P1: b, P2: c})
	if bcDist < closestDistance {
		closestDistance = bcDist
		closestPoints = bcPoints
	}

	caPoints, caDist := ClosestPointsLineVSLine(line, collider.Line{P1: c, P2: a})
	if caDist < closestDistance {
		closestDistance = caDist
		closestPoints = caPoints
	}

	p1Projection, inTriangle := ProjectPointOnTriangle(line.P1, triangle)
	if inTriangle {
		p1Dist := line.P1.Sub(p1Projection).Len()
		if p1Dist < closestDistance {
			closestDistance = p1Dist
			closestPoints = []mgl64.Vec3{line.P1, p1Projection}
		}
	}

	p2Projection, inTriangle := ProjectPointOnTriangle(line.P2, triangle)
	if inTriangle {
		p2Dist := line.P2.Sub(p2Projection).Len()
		if p2Dist < closestDistance {
			closestDistance = p2Dist
			closestPoints = []mgl64.Vec3{line.P2, p2Projection}
		}
	}

	return closestPoints, closestDistance
}

// Real Time Collision Detection - page 149
// Some wacky math stuff.
func ClosestPointsLineVSLine(line1 collider.Line, line2 collider.Line) ([]mgl64.Vec3, float64) {
	p1 := line1.P1
	q1 := line1.P2
	p2 := line2.P1
	q2 := line2.P2

	d1 := q1.Sub(p1)
	d2 := q2.Sub(p2)
	r := p1.Sub(p2)

	a := d1.Dot(d1)
	e := d2.Dot(d2)
	f := d2.Dot(r)

	// check if either or both lines degenerate into points
	if a <= epsilon && e <= epsilon {
		return []mgl64.Vec3{p1, p2}, p1.Sub(p2).Len()
	}

	var s, t float64
	if a <= epsilon {
		// line1 degenerates into a point
		s = 0
		t = mgl64.Clamp(f/e, 0, 1)
	} else {
		c := d1.Dot(r)
		if e <= epsilon {
			// line2 degenerates into a point
			t = 0
			s = mgl64.Clamp(-c/a, 0, 1)
		} else {
			// non-degenerate case
			b := d1.Dot(d2)
			denom := a*e - b*b

			s = 0
			if denom != 0 {
				s = mgl64.Clamp((b*f-c*e)/denom, 0, 1)
			}

			t = (b*s + f) / e

			// If t in [0,1], done. Otherwise, clamp t and recompute s
			if t < 0 {
				t = 0
				s = mgl64.Clamp(-c/a, 0, 1)
			} else if t > 1 {
				t = 1
				s = mgl64.Clamp((b-c)/a, 0, 1)
			}
		}
	}

	c1 := p1.Add(d1.Mul(s))
	c2 := p2.Add(d2.Mul(t))
	return []mgl64.Vec3{c1, c2}, c1.Sub(c2).Len()
}

func ProjectPointOnTriangle(point mgl64.Vec3, triangle collider.Triangle) (mgl64.Vec3, bool) {
	ray := collider.Ray{
		Origin:    point,
		Direction: triangle.Normal.Mul(-1),
	}
	plane := collider.Plane{
		Point:  triangle.Points[0],
		Normal: triangle.Normal,
	}
	intersectionPoint, _ := IntersectRayPlane(ray, plane)
	if intersectionPoint == nil {
		panic("unexpected intersection point to be nil")
	}

	// in point triangle test on page 204
	return *intersectionPoint, PointInTriangle(*intersectionPoint, triangle)
}

// Test if a point is in or on a triangle
func PointInTriangle(point mgl64.Vec3, triangle collider.Triangle) bool {
	// reorient points onto origin based off of point
	a := triangle.Points[0].Sub(point)
	b := triangle.Points[1].Sub(point)
	c := triangle.Points[2].Sub(point)

	u := b.Cross(c)
	v := c.Cross(a)

	if u.Dot(v) < 0 {
		return false
	}

	w := a.Cross(b)

	if u.Dot(w) < 0 {
		return false
	}

	return true
}

func PointInAABB(point mgl64.Vec3, boundingBox *collider.BoundingBox) bool {
	if point.X() < boundingBox.MinVertex.X() || point.X() > boundingBox.MaxVertex.X() {
		return false
	}
	if point.Y() < boundingBox.MinVertex.Y() || point.Y() > boundingBox.MaxVertex.Y() {
		return false
	}
	if point.Z() < boundingBox.MinVertex.Z() || point.Z() > boundingBox.MaxVertex.Z() {
		return false
	}
	return true
}
