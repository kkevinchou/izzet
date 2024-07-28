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

func FindPath(nm *CompiledNavMesh, start, goal mgl64.Vec3) []int {
	tile := nm.Tiles[0]

	_, startPolygon, success := FindNearestPolygon(tile, start)
	if !success {
		fmt.Println("failed to find start poly")
	}
	_, goalPolygon, success := FindNearestPolygon(tile, goal)
	if !success {
		fmt.Println("failed to find goal poly")
	}

	open := gheap.New(Less)
	open.Push(&Node{Polygon: startPolygon, Cost: 0})

	var lastBestNode *Node
	lastBestCost := start.Sub(goal).LenSqr()

	nodeMap := map[int]*Node{}

	for open.Len() > 0 {
		node := open.Pop()
		node.InOpenList = false
		node.InClosedList = true

		polygonIndex := node.Polygon

		if polygonIndex == goalPolygon {
			break
		}

		polygon := tile.Polygons[polygonIndex]

		var g, h float64
		for _, neighborIndex := range polygon.PolyNeighbors {
			if neighborIndex == -1 || (node.Parent != nil && node.Parent.Polygon == neighborIndex) {
				continue
			}
			midpoint, success := GetEdgeMidpoint(node.Polygon, neighborIndex, tile)
			if !success {
				panic("failed to get edge mid point")
			}

			var neighborNode *Node
			if nn, ok := nodeMap[neighborIndex]; ok {
				neighborNode = nn
			} else {
				neighborNode = &Node{
					Position: midpoint,
					Polygon:  neighborIndex,
				}
				nodeMap[neighborIndex] = neighborNode
			}

			if neighborIndex == goalPolygon {
				endCost := neighborNode.Position.Sub(goal).LenSqr()
				g = node.Cost + neighborNode.Position.Sub(node.Position).LenSqr() + endCost
				h = 0
			} else {
				g = node.Cost + neighborNode.Position.Sub(node.Position).LenSqr()
				h = neighborNode.Position.Sub(goal).LenSqr()
			}

			total := g + h

			if neighborNode.InOpenList && total >= neighborNode.Total {
				continue
			}
			if neighborNode.InClosedList && total >= neighborNode.Total {
				continue
			}

			neighborNode.Cost = g
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

			if h < lastBestCost {
				lastBestNode = neighborNode
				lastBestCost = h
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

func GetEdgeMidpoint(from, to int, tile CTile) (mgl64.Vec3, bool) {
	left, right, success := GetPortalVertIndices(from, to, tile)
	if !success {
		return mgl64.Vec3{}, false
	}

	leftVert := tile.Vertices[left]
	rightVert := tile.Vertices[right]

	return leftVert.Add(rightVert).Mul(.5), true
}

func GetPortalVertIndices(from, to int, tile CTile) (int, int, bool) {
	fromPoly := tile.Polygons[from]

	for i, neighborIndex := range fromPoly.PolyNeighbors {
		if neighborIndex == -1 || neighborIndex != to {
			continue
		}

		ni := (i + 1) % len(fromPoly.Vertices)

		left := fromPoly.Vertices[i]
		right := fromPoly.Vertices[ni]

		return left, right, true
	}

	return -1, -1, false
}

func FindNearestPolygon(tile CTile, point mgl64.Vec3) (mgl64.Vec3, int, bool) {
	var nearestDistSq float64 = math.MaxFloat64
	var nearestPoint mgl64.Vec3
	var nearestPoly int

	for i, _ := range tile.Polygons {
		// find neareast point on the polygon
		// height should be taken from the detailed mesh
		pt, success := closestPointOnPoly(tile, i, point)
		if !success {
			continue
		}

		distSq := pt.Sub(point).LenSqr()
		if distSq < nearestDistSq {
			nearestDistSq = distSq
			nearestPoint = pt
			nearestPoly = i
		}
	}

	if nearestDistSq == math.MaxFloat64 {
		return nearestPoint, -1, false
	}

	return nearestPoint, nearestPoly, true
}

func closestPointOnPoly(tile CTile, poly int, point mgl64.Vec3) (mgl64.Vec3, bool) {
	h, success := getPolyHeight(tile, poly, point)
	if success {
		np := point
		np[1] = h
		return np, true
	}

	return point, false
}

func getPolyHeight(tile CTile, poly int, point mgl64.Vec3) (float64, bool) {
	// project point onto polygon
	// early return if it's not within the poly

	if !pointInPoly(tile, poly, point) {
		return -1, false
	}

	dp := tile.CDetailedPolygon[poly]

	for _, tri := range dp.Triangles {
		v0 := tile.CDetailedVertices[poly][tri[0]]
		v1 := tile.CDetailedVertices[poly][tri[1]]
		v2 := tile.CDetailedVertices[poly][tri[2]]

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
