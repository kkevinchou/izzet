package checks

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
)

const (
	epsilon float64 = 0.000001
)

func ProjectPointOnPlane(point mgl64.Vec3, plane collider.Plane) mgl64.Vec3 {
	t := (point.Sub(plane.Point)).Dot(plane.Normal)
	projectedPoint := point.Add(plane.Normal.Mul(-t))
	return projectedPoint
}

func IntersectRayPlane(ray collider.Ray, plane collider.Plane) (mgl64.Vec3, bool) {
	// front determines whether the ray can hit the plane or not
	directionDotNormal := ray.Direction.Dot(plane.Normal)
	if directionDotNormal == 0 {
		return mgl64.Vec3{}, false
	}

	t := (plane.Point.Sub(ray.Origin)).Dot(plane.Normal)
	t /= directionDotNormal

	if t < 0 {
		return mgl64.Vec3{}, false
	}

	intersectionPoint := ray.Origin.Add(ray.Direction.Mul(t))
	return intersectionPoint, true
}

func IntersectRayTriangle(ray collider.Ray, triangle collider.Triangle) (mgl64.Vec3, mgl64.Vec3, bool) {
	nDotDir := triangle.Normal.Dot(ray.Direction)
	if math.Abs(nDotDir) <= 0.001 {
		// ray direction is perpendicular to normal
		return mgl64.Vec3{}, mgl64.Vec3{}, false
	}

	d := triangle.Points[0].Dot(triangle.Normal)
	t := (d - triangle.Normal.Dot(ray.Origin)) / nDotDir
	if t < 0 {
		// don't count plane from behind
		return mgl64.Vec3{}, mgl64.Vec3{}, false
	}

	point := ray.Origin.Add(ray.Direction.Mul(t))

	if PointInTriangle(point, triangle) {
		return point, triangle.Normal, true
	}

	return mgl64.Vec3{}, mgl64.Vec3{}, false
}

func IntersectRayTriMesh(ray collider.Ray, triMesh collider.TriMesh) (mgl64.Vec3, mgl64.Vec3, bool) {
	var minDist float64
	var minPoint mgl64.Vec3
	var minNormal mgl64.Vec3
	var rayHasHit bool

	for _, t := range triMesh.Triangles {
		point, normal, hit := IntersectRayTriangle(ray, t)
		if !hit {
			continue
		}

		if !rayHasHit {
			minDist = ray.Origin.Sub(point).Len()
			minPoint = point
			minNormal = normal
		} else {
			dst := ray.Origin.Sub(point).Len()
			if dst < minDist {
				minDist = dst
				minPoint = point
				minNormal = normal
			}
		}

		rayHasHit = true
	}

	if rayHasHit {
		return minPoint, minNormal, true
	}

	return mgl64.Vec3{}, mgl64.Vec3{}, false
}

func IntersectLineAABB(line collider.Line, bb collider.BoundingBox) (mgl64.Vec3, mgl64.Vec3, bool) {
	tMin := 0.0
	tMax := 1.0

	p1 := line.P1
	p2 := line.P2
	boxMin := bb.MinVertex
	boxMax := bb.MaxVertex

	d := mgl64.Vec3{p2.X() - p1.X(), p2.Y() - p1.Y(), p2.Z() - p1.Z()}

	for i := 0; i < 3; i++ {
		var p0i, di, minI, maxI float64

		switch i {
		case 0:
			p0i = p1.X()
			di = d.X()
			minI = boxMin.X()
			maxI = boxMax.X()
		case 1:
			p0i = p1.Y()
			di = d.Y()
			minI = boxMin.Y()
			maxI = boxMax.Y()
		case 2:
			p0i = p1.Z()
			di = d.Z()
			minI = boxMin.Z()
			maxI = boxMax.Z()
		}

		if di == 0 {
			// Line is parallel to slab
			if p0i < minI || p0i > maxI {
				return mgl64.Vec3{}, mgl64.Vec3{}, false // Outside the box
			}
		} else {
			t1 := (minI - p0i) / di
			t2 := (maxI - p0i) / di
			tEnter := math.Min(t1, t2)
			tExit := math.Max(t1, t2)

			tMin = math.Max(tMin, tEnter)
			tMax = math.Min(tMax, tExit)

			if tMin > tMax {
				return mgl64.Vec3{}, mgl64.Vec3{}, false // No intersection
			}
		}
	}

	// Compute intersection points
	enter := mgl64.Vec3{
		p1.X() + tMin*d.X(),
		p1.Y() + tMin*d.Y(),
		p1.Z() + tMin*d.Z(),
	}

	exit := mgl64.Vec3{
		p1.X() + tMax*d.X(),
		p1.Y() + tMax*d.Y(),
		p1.Z() + tMax*d.Z(),
	}

	return enter, exit, true
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

func ClosestPointRayVsPoint(origin mgl64.Vec3, dir mgl64.Vec3, point mgl64.Vec3) mgl64.Vec3 {
	ac := point.Sub(origin)
	t := ac.Dot(dir) / dir.Dot(dir)

	return origin.Add(dir.Mul(t))
}

// 6 cases
// segment PQ and triangle edge AB,
// segment PQ and triangle edge BC,
// segment PQ and triangle edge CA,
// segment endpoint P and plane of triangle (when P projects inside ABC), and
// segment endpoint Q and plane of triangle (when Q projects inside ABC)

// the first point belongs to the triangle, the second point belongs to the line
func ClosestPointsLineVSTriangle(line collider.Line, triangle collider.Triangle) ([2]mgl64.Vec3, float64) {
	var closestPoints [2]mgl64.Vec3
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
			closestPoints = [2]mgl64.Vec3{line.P1, p1Projection}
		}
	}

	p2Projection, inTriangle := ProjectPointOnTriangle(line.P2, triangle)
	if inTriangle {
		p2Dist := line.P2.Sub(p2Projection).Len()
		if p2Dist < closestDistance {
			closestDistance = p2Dist
			closestPoints = [2]mgl64.Vec3{line.P2, p2Projection}
		}
	}

	return closestPoints, closestDistance
}

// Real Time Collision Detection - page 149
// Some wacky math stuff.
func ClosestPointsLineVSLine(line1 collider.Line, line2 collider.Line) ([2]mgl64.Vec3, float64) {
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
		return [2]mgl64.Vec3{p1, p2}, p1.Sub(p2).Len()
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
	return [2]mgl64.Vec3{c1, c2}, c1.Sub(c2).Len()
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
	projectedPoint := ProjectPointOnPlane(ray.Origin, plane)

	// in point triangle test on page 204
	return projectedPoint, PointInTriangle(projectedPoint, triangle)
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

func ClosestPointsInfiniteLines(p1, q1, p2, q2 mgl64.Vec3) (mgl64.Vec3, mgl64.Vec3, bool) {
	s, t, nonParallel := closestPointsInfiniteLinesMathTest(p1, q1, p2, q2)
	if !nonParallel {
		return mgl64.Vec3{}, mgl64.Vec3{}, nonParallel
	}

	l1 := p1.Add(q1.Sub(p1).Mul(s))
	l2 := p2.Add(q2.Sub(p2).Mul(t))
	return l1, l2, nonParallel
}

// ClosestPointsInfiniteLines finds the closest point between two infinite lines defined by (p1, q1), (p2, q2)
// the boolean return value signals whether the two lines are non-parallel
// Real Time Collision Detection - page 147
func closestPointsInfiniteLinesMathTest(p1, q1, p2, q2 mgl64.Vec3) (float64, float64, bool) {
	r := p1.Sub(p2)
	d1 := q1.Sub(p1)
	d2 := q2.Sub(p2)
	a := d1.Dot(d1)
	b := d1.Dot(d2)
	c := d1.Dot(r)
	e := d2.Dot(d2)
	f := d2.Dot(r)

	d := a*e - (b * b)
	// two lines are parallel
	if d == 0 {
		return 0, 0, false
	}

	s := (b*f - c*e) / d
	t := (a*f - b*c) / d

	return s, t, true
}

// ClosestPointsInfiniteLineVSLine finds the closest point between an infinite line and a line segment (p1, q1), (p2, q2)
// the boolean return value signals whether the two lines are non-parallel
// Real Time Collision Detection - page 147
func ClosestPointsInfiniteLineVSLine(p1, q1, p2, q2 mgl64.Vec3) (mgl64.Vec3, mgl64.Vec3, bool) {
	s, t, nonParallel := closestPointsInfiniteLinesMathTest(p1, q1, p2, q2)
	if !nonParallel {
		return mgl64.Vec3{}, mgl64.Vec3{}, nonParallel
	}

	l1 := p1.Add(q1.Sub(p1).Mul(s))
	l2 := p2.Add(q2.Sub(p2).Mul(t))

	if t > 0 {
		if l2.Sub(p2).Len() > q2.Sub(p2).Len() {
			return l1, q2, nonParallel
		}
		return l1, l2, nonParallel
	}

	return l1, p2, nonParallel
}

func BoundingBoxOverlaps(bb1, bb2 collider.BoundingBox) bool {
	// observer.OnBoundingBoxCheck(e1, e2)

	if bb1.MaxVertex.X() < bb2.MinVertex.X() || bb2.MaxVertex.X() < bb1.MinVertex.X() {
		return false
	}

	if bb1.MaxVertex.Y() < bb2.MinVertex.Y() || bb2.MaxVertex.Y() < bb1.MinVertex.Y() {
		return false
	}

	if bb1.MaxVertex.Z() < bb2.MinVertex.Z() || bb2.MaxVertex.Z() < bb1.MinVertex.Z() {
		return false
	}

	return true
}
