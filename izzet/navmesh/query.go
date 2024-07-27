package navmesh

import (
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

func FindPath(nm *CompiledNavMesh, start, goal mgl64.Vec3, startPolygon, goalPolygon int) *Node {
	// startPolygon := FindClosestPolygon(start, nm)
	// goalPolygon := FindClosestPolygon(goal, nm)

	tile := nm.Tiles[0]

	_, _ = startPolygon, goalPolygon

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
			if neighborIndex == -1 {
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

	PATHPOLYGONS = map[int]bool{}
	n := lastBestNode
	for n != nil {
		PATHPOLYGONS[n.Polygon] = true
		n = n.Parent
	}

	return lastBestNode
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

func FindClosestPolygon(point mgl64.Vec3, nm *CompiledNavMesh) int {
	return -1
}

type NodeHeap []Node

func (h NodeHeap) Len() int           { return len(h) }
func (h NodeHeap) Less(i, j int) bool { return h[i].Cost < h[j].Cost }
func (h NodeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *NodeHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(Node))
}

func (h *NodeHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func Less(n0, n1 *Node) bool {
	return n0.Cost < n1.Cost
}
