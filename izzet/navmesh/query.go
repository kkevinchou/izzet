package navmesh

import (
	"fmt"
	"math"
	"slices"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/gheap"
)

type Path struct {
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

func FindPath(nm *CompiledNavMesh, start, goal mgl64.Vec3) ([]int, []mgl64.Vec3, bool) {
	tile := nm.Tiles[0]

	// start = mgl64.Vec3{-5.953646739493656, 46.86907447625646, -5.199594605599351}
	// goal = mgl64.Vec3{3.081538628009632, 50.35613034511499, -8.687769285772625}

	_, startPolygon, _ := FindNearestPolygon(tile, start)
	_, goalPolygon, _ := FindNearestPolygon(tile, goal)

	startNode := &Node{Polygon: startPolygon, Cost: 0}
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

	PATHVERTICES = FindStraightPath(tile, start, goal, path)

	return path, PATHVERTICES, true
}

func FindStraightPath(tile CTile, start, goal mgl64.Vec3, polyPath []int) []mgl64.Vec3 {
	portalApex := start
	portalLeft := portalApex
	portalRight := portalApex

	var apexIndex, leftIndex, rightIndex int

	closestStart := closestPointOnPolyBoundary(tile, polyPath[0], start)
	closestGoal := closestPointOnPolyBoundary(tile, polyPath[len(polyPath)-1], goal)

	var path []mgl64.Vec3
	path = append(path, closestStart)

	// if len(polyPath) == 1 {
	// 	if _, pidx, success := FindNearestPolygon(tile, goal); success && pidx == polyPath[0] {
	// 		// goal is the same as the start poly, create a straight path and return
	// 		path = append(path, goal)
	// 	} else {
	// 		clippedPoint, success := closestPointOnPoly(tile, polyPath[0], goal)
	// 		if !success {
	// 			panic("WAT")
	// 		}
	// 		path = append(path, clippedPoint)
	// 	}
	// 	// TODO: clip the goal to the poly if it does not lie on the polygon
	// 	return path
	// }

	iterCount := 0
	maxIterCount := 2000

	for i := 0; i < len(polyPath) && iterCount < maxIterCount; i++ {
		iterCount++

		var left, right mgl64.Vec3

		if i+1 < len(polyPath) {
			l, r, success := GetPortalVertIndices(tile, polyPath[i], polyPath[i+1])
			if !success {
				panic(fmt.Sprintf("could not find portal vertices between %d, %d", polyPath[i], polyPath[i+1]))
			}
			left = tile.Vertices[l]
			right = tile.Vertices[r]
		} else {
			left = closestGoal
			right = closestGoal
		}

		// update the right vertex
		if vLeftOn(portalApex, portalRight, right) {
			if vEqual(portalApex, portalRight) || vRight(portalApex, portalLeft, right) {
				// tighten the funnel
				portalRight = right
				rightIndex = i
			} else {
				// right crossed over left, insert left onto the path and restart scan from portal left point
				path = append(path, portalLeft)
				portalApex = portalLeft
				portalRight = portalApex
				apexIndex = leftIndex
				rightIndex = apexIndex
				i = apexIndex
				continue
			}
		}

		// update the right vertex
		if vRightOn(portalApex, portalLeft, left) {
			if vEqual(portalApex, portalLeft) || vLeft(portalApex, portalRight, left) {
				// tighten the funnel
				portalLeft = left
				leftIndex = i
			} else {
				// right crossed over right, insert right onto the path and restart scan from portal right point
				path = append(path, portalRight)
				portalApex = portalRight
				portalLeft = portalApex
				apexIndex = rightIndex
				leftIndex = apexIndex
				i = apexIndex
				continue
			}
		}
	}

	// if _, pidx, success := FindNearestPolygon(tile, goal); success && pidx == polyPath[len(polyPath)-1] {
	// 	path = append(path, goal)
	// } else {
	// 	clippedPoint, success := closestPointOnPoly(tile, polyPath[len(polyPath)-2], goal)
	// 	if !success {
	// 		panic("WAT")
	// 	}
	// 	path = append(path, clippedPoint)
	// }
	path = append(path, closestGoal)

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

	return closestPointOnDetailPolyEdges(tile, poly, point), false
}

func closestPointOnDetailPolyEdges(tile CTile, poly int, point mgl64.Vec3) mgl64.Vec3 {
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

	return vMin.Add(vMax.Sub(vMin).Mul(tMin))
}

func getPolyHeight(tile CTile, poly int, point mgl64.Vec3) (float64, bool) {
	// project point onto polygon
	// early return if it's not within the poly

	if !pointInPoly(tile, poly, point) {
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
