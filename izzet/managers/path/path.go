package path

import (
	"github.com/kkevinchou/izzet/lib/geometry"
	"github.com/kkevinchou/izzet/lib/pathing"
)

func sqWithOffset(size, xOffset, yOffset, zOffset float64) *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{xOffset*size - (size / 2), yOffset, zOffset*size - (size / 2)},
		geometry.Point{xOffset*size - (size / 2), yOffset, zOffset*size + (size / 2)},
		geometry.Point{xOffset*size + (size / 2), yOffset, zOffset*size + (size / 2)},
		geometry.Point{xOffset*size + (size / 2), yOffset, zOffset*size - (size / 2)},
	}
	return geometry.NewPolygon(points)
}

func southRampUpWithOffset(size, elevation, xOffset, yOffset, zOffset float64) *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{xOffset*size - (size / 2), yOffset, zOffset*size - (size / 2)},
		geometry.Point{xOffset*size - (size / 2), yOffset + elevation, zOffset*size + (size / 2)},
		geometry.Point{xOffset*size + (size / 2), yOffset + elevation, zOffset*size + (size / 2)},
		geometry.Point{xOffset*size + (size / 2), yOffset, zOffset*size - (size / 2)},
	}
	return geometry.NewPolygon(points)
}

func funkyShape1() *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{180, 0, 360},
		geometry.Point{180, 0, 420},
		geometry.Point{600, 0, 560},
		geometry.Point{400, 0, 120},
	}
	return geometry.NewPolygon(points)
}

func funkyShape2() *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{500, 0, 50},
		geometry.Point{300, 0, 100},
		geometry.Point{400, 0, 100},
	}
	return geometry.NewPolygon(points)
}

func setupNavMesh() *pathing.NavMesh {
	polygons := []*geometry.Polygon{
		sqWithOffset(5, 0, 0, 0),
		sqWithOffset(5, -1, 0, 0),
		sqWithOffset(5, -1, 0, -1),
		sqWithOffset(5, -1, 0, -2),
		sqWithOffset(5, 0, 0, -2),
		sqWithOffset(5, 1, 0, -2),
		sqWithOffset(5, 2, 0, -2),
		southRampUpWithOffset(5, 4, 2, 0, -1),
		sqWithOffset(5, 2, 4, 0),
		sqWithOffset(5, 3, 4, 0),
		sqWithOffset(5, 4, 4, 0),
		sqWithOffset(5, 4, 4, -1),
		sqWithOffset(5, 4, 4, -2),
		sqWithOffset(5, 3, 4, -2),
		sqWithOffset(5, 2, 4, -2),
		southRampUpWithOffset(5, 4, 2, 4, -1),
		sqWithOffset(5, 2, 8, 0),
		sqWithOffset(5, 3, 8, 0),
		sqWithOffset(5, 4, 8, 0),
		sqWithOffset(5, 4, 8, -1),
		sqWithOffset(5, 4, 8, -2),
		sqWithOffset(5, 3, 8, -2),
		sqWithOffset(5, 2, 8, -2),
		southRampUpWithOffset(5, 4, 2, 8, -1),
		sqWithOffset(5, 2, 12, 0),
		sqWithOffset(5, 3, 12, 0),
		sqWithOffset(5, 4, 12, 0),
		sqWithOffset(5, 4, 12, -1),
		sqWithOffset(5, 4, 12, -2),
		sqWithOffset(5, 3, 12, -2),
		sqWithOffset(5, 2, 12, -2),
	}

	return pathing.ConstructNavMesh(polygons)
}

type Manager struct {
	planner pathing.Planner
	navMesh *pathing.NavMesh
}

func (m *Manager) FindPath(start, goal geometry.Point) []geometry.Point {
	return m.planner.FindPath(start, goal)
}

func (m *Manager) NavMesh() *pathing.NavMesh {
	return m.navMesh
}

func NewManager() *Manager {
	p := pathing.Planner{}
	navMesh := setupNavMesh()
	p.SetNavMesh(navMesh)
	return &Manager{planner: p, navMesh: navMesh}
}
