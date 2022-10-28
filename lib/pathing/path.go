package pathing

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/lib/geometry"
	utils "github.com/kkevinchou/izzet/lib/libutils"
)

func getPoints(path []NavNode) []geometry.Point {
	points := make([]geometry.Point, len(path))
	for i, node := range path {
		points[i] = node.Point
	}
	return points
}

const (
	floatEpsilon = float64(0.001) * float64(0.001)
)

type Planner struct {
	navmesh *NavMesh
}

func (p *Planner) SetNavMesh(navmesh *NavMesh) {
	p.navmesh = navmesh
}

// FindPath finds a path from start to goal. The path returned does not include the start node.
func (p *Planner) FindPath(start geometry.Point, goal geometry.Point) []geometry.Point {
	if start == goal {
		return []geometry.Point{start, goal}
	}

	roughPath := p.findPath(start, goal)
	if roughPath == nil {
		return nil
	}

	if len(roughPath) >= 3 {
		portals := p.findPortals(roughPath)
		return smoothPath(portals)
	}

	return getPoints(roughPath)
}

func (p *Planner) findPortals(path []NavNode) []Portal {
	if len(path) == 0 {
		return nil
	}

	portals := []Portal{Portal{Point1: path[0].Point, Point2: path[0].Point}}
	prevPolygon := path[0].Polygon

	for _, node := range path {
		if node.Polygon != prevPolygon {
			portals = append(portals, p.navmesh.polyPairToPortal[prevPolygon][node.Polygon])
			prevPolygon = node.Polygon
		}
	}

	finalPoint := path[len(path)-1].Point
	portals = append(portals, Portal{Point1: finalPoint, Point2: finalPoint})

	return portals
}

func (p *Planner) findPath(start geometry.Point, goal geometry.Point) []NavNode {
	// Initialize
	frontier := utils.NewPriorityQueue()
	cameFrom := map[NavNode]NavNode{}
	costSoFar := map[NavNode]float64{}

	startPolygonFound := false
	goalPolygonFound := false

	var startNode, goalNode NavNode

	// Find which polygon our start node lies in
	for _, polygon := range p.navmesh.Polygons() {
		if startPolygonFound && goalPolygonFound {
			break
		}

		if !startPolygonFound && polygon.ContainsPoint(start) {
			startNode = NavNode{Point: start, Polygon: polygon}
			startPolygonFound = true
			for _, point := range polygon.Points() {
				node := NavNode{Point: point, Polygon: polygon}
				cost := point.Vector3().Sub(start.Vector3()).Len()
				cameFrom[node] = startNode
				costSoFar[node] = cost

				// Initialize the frontier with each of the neighbors
				// of the start node within the polygon
				frontier.Push(node, cost)
			}
		}

		if !goalPolygonFound && polygon.ContainsPoint(goal) {
			goalNode = NavNode{Point: goal, Polygon: polygon}
			goalPolygonFound = true
		}
	}

	// If we couldn't find the start or goal polygon, abort
	if !startPolygonFound || !goalPolygonFound {
		return nil
	}

	// If we have a direct path from start to goal, return it
	if startNode.Polygon == goalNode.Polygon {
		return []NavNode{startNode, goalNode}
	}

	// Set the goal node as the neighbor of each node in the goal polygon
	goalNeighbors := map[NavNode]bool{}
	for _, point := range goalNode.Polygon.Points() {
		node := NavNode{Point: point, Polygon: goalNode.Polygon}
		goalNeighbors[node] = true
	}

	explored := map[NavNode]bool{}

	// Start searching for a path!
	for !frontier.Empty() {
		current := frontier.Pop().(NavNode)

		if current == goalNode {
			break
		}

		explored[current] = true

		neighbors := p.navmesh.Neighbors(current)

		// Append the goal to the list of neighbors if the current point is a neighbor
		// of the goal point
		if _, ok := goalNeighbors[current]; ok {
			neighbors = append(neighbors, goalNode)
		}

		for _, neighbor := range neighbors {
			if _, ok := explored[neighbor]; ok {
				continue
			}

			// Overwrite the cost to reach the neighbor if the cost is better than
			// what we previously recorded (or if we haven't recorded a cost yet)
			newCost := costSoFar[current] + p.navmesh.Cost(current.Point, neighbor.Point)
			if cost, ok := costSoFar[neighbor]; !ok || newCost < cost {
				costSoFar[neighbor] = newCost
				frontier.Push(neighbor, newCost+p.navmesh.Cost(goalNode.Point, neighbor.Point))
				cameFrom[neighbor] = current
			}
		}
	}

	if _, ok := cameFrom[goalNode]; !ok {
		// Could not find a path to the goal node
		return nil
	}

	path := []NavNode{}
	pathNode := goalNode

	for {
		path = append(path, pathNode)
		if pathNode == startNode {
			break
		}
		pathNode = cameFrom[pathNode]
	}

	reversePath := make([]NavNode, len(path))
	for i := 0; i < len(path); i++ {
		reversePath[len(path)-1-i] = path[i]
	}

	return reversePath
}

func orderPortalPoints(portals []Portal) []geometry.Point {
	portalPoints := []geometry.Point{portals[0].Point1, portals[0].Point2}
	prevLeft := portals[0].Point1
	prevRight := portals[0].Point2

	for i := 1; i < len(portals); i++ {
		nextLeft := portals[i].Point1
		nextRight := portals[i].Point2

		leftVec := nextLeft.Vector3().Sub(prevLeft.Vector3())
		rightVec := nextRight.Vector3().Sub(prevRight.Vector3())

		// TODO: handle Cross
		if utils.Cross2D(rightVec, leftVec) > 0 {
			nextLeft, nextRight = nextRight, nextLeft
		}
		// TODO: handle where they're == 0

		portalPoints = append(portalPoints, nextLeft)
		portalPoints = append(portalPoints, nextRight)

		prevLeft, prevRight = nextLeft, nextRight
	}

	return portalPoints
}

// Returns true if v is to left of reference
func vecOnLeft(reference, v mgl64.Vec3) bool {
	return utils.Cross2D(reference, v) < floatEpsilon
	// return reference.Cross(v) < 0
}

// Returns true if v is to the right of reference
func vecOnRight(reference, v mgl64.Vec3) bool {
	return utils.Cross2D(reference, v) > -1*floatEpsilon
	// return reference.Cross(v) > 0
}

func smoothPath(unorderedPortals []Portal) []geometry.Point {
	portalPoints := orderPortalPoints(unorderedPortals)

	// This algorithm was retrieved online but a confusing note:
	// lastValidRightIndex actually represent "left" index
	//
	// lastValidRightIndex represents the left index of the last valid
	// right index.  These indexes are used purely to reset the apex
	// at the correct point

	lastValidLeftIndex := 0
	lastValidRightIndex := 0

	apex := portalPoints[0]
	portalLeft := apex
	portalRight := apex

	contactPoints := []geometry.Point{apex}

	for i := 2; i < len(portalPoints); i += 2 {
		leftPoint := portalPoints[i]
		rightPoint := portalPoints[i+1]

		leftVec := leftPoint.Vector3().Sub(apex.Vector3())
		rightVec := rightPoint.Vector3().Sub(apex.Vector3())
		lastValidLeftVec := portalLeft.Vector3().Sub(apex.Vector3())
		lastValidRightVec := portalRight.Vector3().Sub(apex.Vector3())

		// Left side of funnel
		// The leftVec is to the right of lastValidLeftVec, so we
		// shrink the funnel
		if vecOnLeft(leftVec, lastValidLeftVec) {
			if (portalLeft == apex) || !vecOnRight(lastValidRightVec, leftVec) {
				portalLeft = leftPoint
				lastValidLeftIndex = i
			} else {
				// If the new leftVec is to the right of the last valid
				// right vec, we set the new apex
				apex = portalRight
				portalLeft = apex
				if contactPoints[len(contactPoints)-1] != apex {
					contactPoints = append(contactPoints, apex)
				}

				lastValidLeftIndex = lastValidRightIndex
				i = lastValidRightIndex
				continue
			}
		}

		// Right side of funnel
		if vecOnRight(rightVec, lastValidRightVec) {
			if (portalRight == apex) || !vecOnLeft(lastValidLeftVec, rightVec) {
				portalRight = rightPoint
				lastValidRightIndex = i
			} else {
				apex = portalLeft
				portalRight = apex
				if contactPoints[len(contactPoints)-1] != apex {
					contactPoints = append(contactPoints, apex)
				}

				lastValidRightIndex = lastValidLeftIndex
				i = lastValidLeftIndex
				continue
			}
		}
	}

	if contactPoints[len(contactPoints)-1] != portalPoints[len(portalPoints)-1] {
		contactPoints = append(contactPoints, portalPoints[len(portalPoints)-1])
	}

	return contactPoints
}
