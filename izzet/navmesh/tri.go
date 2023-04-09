package navmesh

import "github.com/go-gl/mathgl/mgl64"

func LineSegmentAABBIntersection(lineStart, lineEnd, aabbMin, aabbMax mgl64.Vec3) bool {
	// Calculate the direction vector of the line segment
	direction := lineEnd.Sub(lineStart)

	// Calculate the minimum and maximum values of t for each of the three dimensions
	tMinX := (aabbMin.X() - lineStart.X()) / direction.X()
	tMaxX := (aabbMax.X() - lineStart.X()) / direction.X()
	tMinY := (aabbMin.Y() - lineStart.Y()) / direction.Y()
	tMaxY := (aabbMax.Y() - lineStart.Y()) / direction.Y()
	tMinZ := (aabbMin.Z() - lineStart.Z()) / direction.Z()
	tMaxZ := (aabbMax.Z() - lineStart.Z()) / direction.Z()

	// Calculate the largest minimum value and the smallest maximum value for all three dimensions
	tMin := max(max(min(tMinX, tMaxX), min(tMinY, tMaxY)), min(tMinZ, tMaxZ))
	tMax := min(min(max(tMinX, tMaxX), max(tMinY, tMaxY)), max(tMinZ, tMaxZ))

	// Check if the line segment intersects the AABB
	if tMax < 0 || tMin > tMax {
		return false
	}
	return true
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
