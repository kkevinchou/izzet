package navmesh

import (
	"fmt"
	"math"
	"slices"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/gheap"
)

type pathPortal struct {
	Left        mgl64.Vec3
	Right       mgl64.Vec3
	ProjectPoly int
	End         bool
}

type Node struct {
	Cost     float64
	Total    float64
	Polygon  int
	Parent   *Node
	Position mgl64.Vec3

	InOpenList   bool
	InClosedList bool
}

var PATHPOLYGONS map[int]bool
var PATHVERTICES []mgl64.Vec3

// FindPath returns a list of polygons through which it is possible to path from start to goal
func FindPath(nm *CompiledNavMesh, start, goal mgl64.Vec3) []int {
	tile := nm.Tiles[0]

	_, startPolygon, _ := FindNearestPolygon(tile, start)
	_, goalPolygon, _ := FindNearestPolygon(tile, goal)

	startNode := &Node{Polygon: startPolygon, Cost: 0, Position: start}
	lastBestNode := startNode
	lastBestCost := start.Sub(goal).Len()

	open := gheap.New(Less)
	open.Push(startNode)

	nodeMap := map[int]*Node{}

	for open.Len() > 0 {
		node := open.Pop()
		node.InOpenList = false
		node.InClosedList = true

		polygonIndex := node.Polygon

		if polygonIndex == goalPolygon {
			lastBestNode = node
			break
		}

		polygon := tile.Polygons[polygonIndex]

		var cost, heuristic float64
		for _, neighborIndex := range polygon.PolyNeighbors {
			if neighborIndex == -1 || (node.Parent != nil && node.Parent.Polygon == neighborIndex) {
				continue
			}

			var neighborNode *Node
			if nn, ok := nodeMap[neighborIndex]; ok {
				neighborNode = nn
			} else {
				midpoint, success := GetEdgeMidpoint(tile, node.Polygon, neighborIndex)
				if !success {
					panic("failed to get edge mid point")
				}
				neighborNode = &Node{
					Position: midpoint,
					Polygon:  neighborIndex,
				}
				nodeMap[neighborIndex] = neighborNode
			}

			if neighborIndex == goalPolygon {
				endCost := neighborNode.Position.Sub(goal).Len()
				cost = node.Cost + neighborNode.Position.Sub(node.Position).Len() + endCost
				heuristic = 0
			} else {
				cost = node.Cost + neighborNode.Position.Sub(node.Position).Len()
				heuristic = neighborNode.Position.Sub(goal).Len()
			}

			total := cost + heuristic

			if neighborNode.InOpenList && total >= neighborNode.Total {
				continue
			}
			if neighborNode.InClosedList && total >= neighborNode.Total {
				continue
			}

			neighborNode.Cost = cost
			neighborNode.Parent = node
			neighborNode.Total = total
			neighborNode.InClosedList = false

			if neighborNode.InOpenList {
				for i, node := range open.Slice {
					if node.Polygon == neighborIndex {
						open.Fix(i)
						break
					}
				}
			} else {
				neighborNode.InOpenList = true
				open.Push(neighborNode)
			}

			if heuristic < lastBestCost {
				lastBestNode = neighborNode
				lastBestCost = heuristic
			}
		}
	}

	var path []int
	n := lastBestNode
	for n != nil {
		path = append(path, n.Polygon)
		n = n.Parent
	}
	slices.Reverse(path)

	return path
}

// FindStraightPath takes polyPath which is a list of polygons and runs the funnel
// algorithm along the portals between each polygon.
//
// the funnel is the actively managed funnel that we are attempting tho tighten
// portals are the shared edge connection between two polygons
func FindStraightPath(tile CTile, start, goal mgl64.Vec3, polyPath []int) []mgl64.Vec3 {
	if len(polyPath) == 0 {
		return nil
	}

	closestStart, _ := closestPointOnPoly(tile, polyPath[0], start)
	closestGoal, _ := closestPointOnPoly(tile, polyPath[len(polyPath)-1], goal)

	portals := buildPathPortals(tile, polyPath, closestGoal)

	funnelApex := closestStart
	funnelLeft := funnelApex
	funnelRight := funnelApex

	apexIndex := -1
	leftIndex := -1
	rightIndex := -1

	path := []mgl64.Vec3{closestStart}

	iterCount := 0
	maxIterCount := 2000

	for i := 0; i < len(portals) && iterCount < maxIterCount; i++ {
		iterCount++

		portal := portals[i]

		if vLeftOn(funnelApex, funnelRight, portal.Right) {
			if vEqual(funnelApex, funnelRight) || vRight(funnelApex, funnelLeft, portal.Right) {
				// the portal's right vertex lies within the funnel, update the funnel's right
				funnelRight = portal.Right
				rightIndex = i
			} else {
				// the portal's right vertex collapses the funnel and we've discovered a turning update the apex
				if appendPortalPoint(&path, tile, portals[leftIndex], funnelLeft) {
					return path
				}

				// collapse the funnel into a new singular apex
				funnelApex = funnelLeft
				funnelLeft = funnelApex
				funnelRight = funnelApex

				// reset the funnel algorithm from the new apex
				apexIndex = leftIndex
				leftIndex = apexIndex
				rightIndex = apexIndex
				i = apexIndex

				continue
			}
		}

		if vRightOn(funnelApex, funnelLeft, portal.Left) {
			if vEqual(funnelApex, funnelLeft) || vLeft(funnelApex, funnelRight, portal.Left) {
				// the portal's left vertex lies within the funnel, update the funnel's left
				funnelLeft = portal.Left
				leftIndex = i
			} else {
				// the portal's left vertex collapses the funnel and we've discovered a turning point update the apex
				if appendPortalPoint(&path, tile, portals[rightIndex], funnelRight) {
					return path
				}

				// collapse the funnel into a new singular apex
				funnelApex = funnelRight
				funnelLeft = funnelApex
				funnelRight = funnelApex

				// reset the funnel algorithm from the new apex
				apexIndex = rightIndex
				leftIndex = apexIndex
				rightIndex = apexIndex
				i = apexIndex

				continue
			}
		}
	}

	appendPoint(&path, closestGoal)

	return path
}

func vEqual(a, b mgl64.Vec3) bool {
	threshold := (1.0 / 16384.0) * (1.0 / 16384.0)
	return a.Sub(b).LenSqr() < threshold
}

// returns true if c_a is to the left of b_a
func vLeftOn(a, b, c mgl64.Vec3) bool {
	return vArea2D(a, b, c) <= 0
}

func vLeft(a, b, c mgl64.Vec3) bool {
	return vArea2D(a, b, c) < 0
}

func vRightOn(a, b, c mgl64.Vec3) bool {
	return vArea2D(a, b, c) >= 0
}

func vRight(a, b, c mgl64.Vec3) bool {
	return vArea2D(a, b, c) > 0
}
func vArea2D(a, b, c mgl64.Vec3) float64 {
	p := (b.X() - a.X()) * (c.Z() - a.Z())
	q := (c.X() - a.X()) * (b.Z() - a.Z())
	value := p - q
	return value
}
func GetEdgeMidpoint(tile CTile, from, to int) (mgl64.Vec3, bool) {
	left, right, success := GetPortalVertIndices(tile, from, to)
	if !success {
		return mgl64.Vec3{}, false
	}

	leftVert := tile.Vertices[left]
	rightVert := tile.Vertices[right]

	return leftVert.Add(rightVert).Mul(.5), true
}

func GetPortalVertIndices(tile CTile, from, to int) (int, int, bool) {
	fromPoly := tile.Polygons[from]

	for i, neighborIndex := range fromPoly.PolyNeighbors {
		if neighborIndex == -1 || neighborIndex != to {
			continue
		}

		ni := (i + 1) % len(fromPoly.Vertices)

		left := fromPoly.Vertices[ni]
		right := fromPoly.Vertices[i]

		return left, right, true
	}

	return -1, -1, false
}

func FindNearestPolygon(tile CTile, point mgl64.Vec3) (mgl64.Vec3, int, bool) {
	var nearestDistSq float64 = math.MaxFloat64
	var nearestPoint mgl64.Vec3
	var nearestPoly int
	var overPoly bool

	for i, _ := range tile.Polygons {
		// find neareast point on the polygon
		// height should be taken from the detailed mesh
		pt, op := closestPointOnPoly(tile, i, point)
		distSq := pt.Sub(point).LenSqr()

		if distSq < nearestDistSq {
			nearestDistSq = distSq
			nearestPoint = pt
			nearestPoly = i
			overPoly = op
		}
	}

	return nearestPoint, nearestPoly, overPoly
}

// closestPointOnPolyBoundary is faster than closestPointOnPoly, but does not return detailed heights
// on the polygon - if a point is within the x-z bounds, it returns the point, unmodified
// if the point is outside of the poly, it returns the closest point on the boundary of the poly
func closestPointOnPolyBoundary(tile CTile, poly int, point mgl64.Vec3) mgl64.Vec3 {
	if pointInPoly(tile, poly, point) {
		return point
	}
	return closestPointOnPolyEdges(tile, poly, point)
}

func closestPointOnPolyEdges(tile CTile, poly int, point mgl64.Vec3) mgl64.Vec3 {
	var dmin float64 = math.MaxFloat64
	var tMin float64
	var vMin, vMax mgl64.Vec3
	verts := tile.Polygons[poly].Vertices
	for i, j := 0, len(verts)-1; i < len(verts); j, i = i, i+1 {
		v0 := tile.Vertices[verts[j]]
		v1 := tile.Vertices[verts[i]]
		d, t := distancePtSeg2Df(point.X(), point.Z(), v0.X(), v0.Z(), v1.X(), v1.Z())
		if d < dmin {
			dmin = d
			tMin = t
			vMin, vMax = v0, v1
		}
	}

	return vMin.Add(vMax.Sub(vMin).Mul(tMin))
}

func closestPointOnPoly(tile CTile, poly int, point mgl64.Vec3) (mgl64.Vec3, bool) {
	h, success := getPolyHeight(tile, poly, point)
	if success {
		np := point
		np[1] = h
		return np, true
	}

	if !hasDetailedPoly(tile, poly) {
		return closestPointOnPolyEdges(tile, poly, point), false
	}

	return closestPointOnDetailPolyEdges(tile, poly, point), false
}

func closestPointOnDetailPolyEdges(tile CTile, poly int, point mgl64.Vec3) mgl64.Vec3 {
	if !hasDetailedPoly(tile, poly) {
		return closestPointOnPolyEdges(tile, poly, point)
	}

	dp := tile.DetailedPolygon[poly]
	var dmin float64 = math.MaxFloat64
	var tMin float64
	var vMin, vMax mgl64.Vec3

	for _, tri := range dp.Triangles {
		v0 := tile.DetailedVertices[poly][tri.Vertices[0]]
		v1 := tile.DetailedVertices[poly][tri.Vertices[1]]
		v2 := tile.DetailedVertices[poly][tri.Vertices[2]]

		var d, t float64
		if tri.OnHull[0] {
			d, t = distancePtSeg2Df(point.X(), point.Z(), v0.X(), v0.Z(), v1.X(), v1.Z())
			if d < dmin {
				dmin = d
				tMin = t
				vMin, vMax = v0, v1
			}
		}

		if tri.OnHull[1] {
			d, t = distancePtSeg2Df(point.X(), point.Z(), v1.X(), v1.Z(), v2.X(), v2.Z())
			if d < dmin {
				dmin = d
				tMin = t
				vMin, vMax = v1, v2
			}
		}

		if tri.OnHull[2] {
			d, t = distancePtSeg2Df(point.X(), point.Z(), v2.X(), v2.Z(), v0.X(), v0.Z())
			if d < dmin {
				dmin = d
				tMin = t
				vMin, vMax = v2, v0
			}
		}
	}

	if dmin == math.MaxFloat64 {
		return closestPointOnPolyEdges(tile, poly, point)
	}

	return vMin.Add(vMax.Sub(vMin).Mul(tMin))
}

func getPolyHeight(tile CTile, poly int, point mgl64.Vec3) (float64, bool) {
	if !hasDetailedPoly(tile, poly) {
		return -1, false
	}

	dp := tile.DetailedPolygon[poly]

	for _, tri := range dp.Triangles {
		v0 := tile.DetailedVertices[poly][tri.Vertices[0]]
		v1 := tile.DetailedVertices[poly][tri.Vertices[1]]
		v2 := tile.DetailedVertices[poly][tri.Vertices[2]]

		if height, success := closestHeightOnTriangle(point, v0, v1, v2); success {
			return height, true
		}
	}

	// TODO: check against detailed mesh edges

	return -1, false
}

func hasDetailedPoly(tile CTile, poly int) bool {
	return poly >= 0 &&
		poly < len(tile.DetailedPolygon) &&
		poly < len(tile.DetailedVertices) &&
		len(tile.DetailedPolygon[poly].Triangles) > 0 &&
		len(tile.DetailedVertices[poly]) > 0
}

func closestHeightOnTriangle(p, a, b, c mgl64.Vec3) (float64, bool) {
	epsilon := 1e-6
	v0 := c.Sub(a)
	v1 := b.Sub(a)
	v2 := p.Sub(a)

	// compute scaled barycentric coordinates
	denom := v0.X()*v1.Z() - v0.Z()*v1.X()
	if math.Abs(denom) < epsilon {
		return -1, false
	}

	u := v1.Z()*v2.X() - v1.X()*v2.Z()
	v := v0.X()*v2.Z() - v0.Z()*v2.X()

	if denom < 0 {
		denom = -denom
		u = -u
		v = -v
	}

	// if the point lies within the triangle, return the interpolated y value
	if u >= 0 && v >= 0 && (u+v) <= denom {
		h := a.Y() + (v0.Y()*u+v1.Y()*v)/denom
		return h, true
	}

	return -1, false
}

func pointInPoly(tile CTile, poly int, point mgl64.Vec3) bool {
	verts := tile.Vertices
	vertIndices := tile.Polygons[poly].Vertices
	n := len(vertIndices)
	c := false

	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		vi := verts[vertIndices[i]]
		vj := verts[vertIndices[j]]
		if ((vi.Z() > point.Z()) != (vj.Z() > point.Z())) && (point.X() < (vj.X()-vi.X())*(point.Z()-vi.Z())/(vj.Z()-vi.Z())+vi.X()) {
			c = !c
		}
	}

	return c
}

func Less(n0, n1 *Node) bool {
	return n0.Cost < n1.Cost
}

func buildPathPortals(tile CTile, polyPath []int, closestGoal mgl64.Vec3) []pathPortal {
	portals := make([]pathPortal, 0, len(polyPath))
	for i := 1; i < len(polyPath); i++ {
		l, r, success := GetPortalVertIndices(tile, polyPath[i-1], polyPath[i])
		if !success {
			panic(fmt.Sprintf("could not find portal vertices between %d, %d", polyPath[i-1], polyPath[i]))
		}
		portals = append(portals, pathPortal{
			Left:        tile.Vertices[l],
			Right:       tile.Vertices[r],
			ProjectPoly: polyPath[i],
		})
	}

	portals = append(portals, pathPortal{
		Left:        closestGoal,
		Right:       closestGoal,
		ProjectPoly: polyPath[len(polyPath)-1],
		End:         true,
	})

	return portals
}

func appendPortalPoint(path *[]mgl64.Vec3, tile CTile, portal pathPortal, point mgl64.Vec3) bool {
	if portal.End {
		appendPoint(path, point)
		return true
	}

	appendPoint(path, projectPathPoint(tile, portal.ProjectPoly, point))
	return false
}

func appendPoint(path *[]mgl64.Vec3, point mgl64.Vec3) {
	if len(*path) > 0 && vEqual((*path)[len(*path)-1], point) {
		return
	}
	*path = append(*path, point)
}

// project the waypoint onto a detailed polygon if possible to extract a more detailed y value
func projectPathPoint(tile CTile, poly int, point mgl64.Vec3) mgl64.Vec3 {
	if poly < 0 || poly >= len(tile.Polygons) {
		return point
	}

	projected, _ := closestPointOnPoly(tile, poly, point)
	return projected
}
