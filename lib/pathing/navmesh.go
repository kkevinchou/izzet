package pathing

import (
	"fmt"

	"github.com/kkevinchou/izzet/lib/geometry"
)

type NavNode struct {
	Point   geometry.Point
	Polygon *geometry.Polygon
}

// Portals are edges that join two polygons
type Portal struct {
	Point1 geometry.Point
	Point2 geometry.Point
}

func (p Portal) String() string {
	return fmt.Sprintf("P{%v, %v}", p.Point1, p.Point2)
}

type NavMesh struct {
	neighbors        map[NavNode][]NavNode
	costs            map[NavNode]map[NavNode]float64
	polygons         []*geometry.Polygon
	portalToPolygons map[Portal][]*geometry.Polygon
	polyPairToPortal map[*geometry.Polygon]map[*geometry.Polygon]Portal

	*RenderComponent
}

func ConstructNavMesh(polygons []*geometry.Polygon) *NavMesh {
	navmesh := &NavMesh{
		neighbors:        map[NavNode][]NavNode{},
		costs:            map[NavNode]map[NavNode]float64{},
		portalToPolygons: map[Portal][]*geometry.Polygon{},
		polyPairToPortal: map[*geometry.Polygon]map[*geometry.Polygon]Portal{},
	}

	for _, polygon := range polygons {
		navmesh.AddPolygon(polygon)
	}

	navmesh.RenderComponent = &RenderComponent{
		RenderData: &NavMeshRenderData{
			ID:      "tile",
			Visible: true,
		},
	}

	return navmesh
}
func (nm *NavMesh) Polygons() []*geometry.Polygon {
	return nm.polygons
}

func (nm *NavMesh) AddPolygon(polygon *geometry.Polygon) {
	nm.polygons = append(nm.polygons, polygon)
	for i, point1 := range polygon.Points() {
		for j, point2 := range polygon.Points() {
			// Avoid processing the same two pairs of indicies
			if i >= j {
				continue
			}

			navNode1 := NavNode{Point: point1, Polygon: polygon}
			navNode2 := NavNode{Point: point2, Polygon: polygon}

			// TODO: handle duplicate edges being added
			nm.addEdge(navNode1, navNode2)
			nm.addEdge(navNode2, navNode1)

			// Make a deterministically ordered Portal for consistent lookups
			portal := Portal{Point1: point1, Point2: point2}
			if point1[0] != point2[0] {
				if point1[0] > point2[0] {
					portal = Portal{Point1: point1, Point2: point2}
				} else {
					portal = Portal{Point1: point2, Point2: point1}
				}
			} else if point1[1] != point2[1] {
				if point1[1] > point2[1] {
					portal = Portal{Point1: point1, Point2: point2}
				} else {
					portal = Portal{Point1: point2, Point2: point1}
				}
			} else if point1[2] != point2[2] {
				if point1[2] > point2[2] {
					portal = Portal{Point1: point1, Point2: point2}
				} else {
					portal = Portal{Point1: point2, Point2: point1}
				}
			}

			// Found one half of the portal, complete the other half
			if len(nm.portalToPolygons[portal]) == 1 {
				polyWithSharedPortal := nm.portalToPolygons[portal][0]
				if polyWithSharedPortal != polygon {
					// Set up the neighbors
					if _, ok := nm.polyPairToPortal[polygon]; !ok {
						nm.polyPairToPortal[polygon] = map[*geometry.Polygon]Portal{}
					}
					if _, ok := nm.polyPairToPortal[polyWithSharedPortal]; !ok {
						nm.polyPairToPortal[polyWithSharedPortal] = map[*geometry.Polygon]Portal{}
					}

					nm.polyPairToPortal[polygon][polyWithSharedPortal] = portal
					nm.polyPairToPortal[polyWithSharedPortal][polygon] = portal

					// Set points that lie on the same portal to be neighbors to one another

					navNode1 := NavNode{Point: point1, Polygon: polygon}
					otherNavNode1 := NavNode{Point: point1, Polygon: polyWithSharedPortal}

					nm.neighbors[navNode1] = append(
						nm.neighbors[navNode1],
						otherNavNode1,
					)

					nm.neighbors[otherNavNode1] = append(
						nm.neighbors[otherNavNode1],
						navNode1,
					)

					navNode2 := NavNode{Point: point2, Polygon: polygon}
					otherNavNode2 := NavNode{Point: point2, Polygon: polyWithSharedPortal}

					nm.neighbors[navNode2] = append(
						nm.neighbors[navNode2],
						otherNavNode2,
					)

					nm.neighbors[otherNavNode2] = append(
						nm.neighbors[otherNavNode2],
						navNode2,
					)
				}
			}
			nm.portalToPolygons[portal] = append(nm.portalToPolygons[portal], polygon)
		}
	}
}

func (nm *NavMesh) addEdge(from, to NavNode) {
	if _, ok := nm.neighbors[from]; !ok {
		nm.neighbors[from] = []NavNode{}
	}
	nm.neighbors[from] = append(nm.neighbors[from], to)
}

func (nm *NavMesh) Neighbors(point NavNode) []NavNode {
	if neighbors, ok := nm.neighbors[point]; ok {
		return copyPointList(neighbors)
	}
	return []NavNode{}
}

func (nm *NavMesh) Cost(from, to geometry.Point) float64 {
	v1 := from.Vector3()
	v2 := to.Vector3()

	var length float64
	// TODO: do i need to do this equality check? seems like .Length
	// will already return 0
	if (v1[0] == v2[0]) && (v1[1] == v2[1]) && (v1[2] == v2[2]) {
		length = 0
	} else {
		length = v1.Sub(v2).Len()
	}

	return length
}

func copyPointList(points []NavNode) []NavNode {
	return append([]NavNode{}, points...)
}
