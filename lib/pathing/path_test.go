package pathing

import (
	"testing"

	"github.com/kkevinchou/izzet/lib/geometry"
)

func tri1() *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{11, 0, 4},
		geometry.Point{13, 0, 10},
		geometry.Point{17, 0, 8},
	}
	return geometry.NewPolygon(points)
}

func tri2() *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{13, 0, 10},
		geometry.Point{12, 0, 13},
		geometry.Point{17, 0, 8},
	}
	return geometry.NewPolygon(points)
}

func tri3() *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{17, 0, 8},
		geometry.Point{12, 0, 13},
		geometry.Point{21, 0, 7},
	}
	return geometry.NewPolygon(points)
}

func tri4() *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{17, 0, 2},
		geometry.Point{17, 0, 8},
		geometry.Point{21, 0, 7},
	}
	return geometry.NewPolygon(points)
}

func TestWithNewApex(t *testing.T) {
	polygons := []*geometry.Polygon{
		tri1(),
		tri2(),
		tri3(),
		tri4(),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{13, 0, 7}, geometry.Point{18, 0, 5})
	expectedPath := []geometry.Point{geometry.Point{13, 0, 7}, geometry.Point{17, 0, 8}, geometry.Point{18, 0, 5}}
	assertPathEq(t, expectedPath, path)
}

func TestSmoothing(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithXOffset(0),
		sqWithXOffset(6),
		sqWithXOffset(12),
		sqWithXOffset(18),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{1, 0, 1}, geometry.Point{17, 0, 5})
	expectedPath := []geometry.Point{geometry.Point{1, 0, 1}, geometry.Point{17, 0, 5}}
	assertPathEq(t, expectedPath, path)
}

// X X X
//     X
//     X X
func TestTwoApexes(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(30, 0, 0),
		sqWithOffset(30, 1, 0),
		sqWithOffset(30, 2, 0),
		sqWithOffset(30, 2, 1),
		sqWithOffset(30, 2, 2),
		sqWithOffset(30, 3, 2),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{0, 0, 0}, geometry.Point{110, 0, 69})
	expectedPath := []geometry.Point{geometry.Point{0, 0}, geometry.Point{60, 0, 30}, geometry.Point{90, 0, 60}, geometry.Point{110, 0, 69}}
	assertPathEq(t, expectedPath, path)
}

func TestStartNodeOverlapsNode(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(30, 0, 0),
		sqWithOffset(30, 1, 0),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{0, 0, 0}, geometry.Point{50, 0, 20})
	expectedPath := []geometry.Point{geometry.Point{0, 0, 0}, geometry.Point{50, 0, 20}}
	assertPathEq(t, expectedPath, path)
}

func TestGoalNodeOverlapsNode(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(30, 0, 0),
		sqWithOffset(30, 1, 0),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{1, 0, 1}, geometry.Point{30, 0, 30})
	expectedPath := []geometry.Point{geometry.Point{1, 0, 1}, geometry.Point{30, 0, 30}}
	assertPathEq(t, expectedPath, path)
}

func TestStartAndGoalNodeOverlapsNode(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(30, 0, 0),
		sqWithOffset(30, 1, 0),
		sqWithOffset(30, 1, 1),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{0, 0, 0}, geometry.Point{30, 0, 60})
	expectedPath := []geometry.Point{geometry.Point{0, 0, 0}, geometry.Point{30, 0, 30}, geometry.Point{30, 0, 60}}
	assertPathEq(t, expectedPath, path)
}

func TestReverseC(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(60, 0, 0),
		sqWithOffset(60, 1, 0),
		sqWithOffset(60, 1, 1),
		sqWithOffset(60, 1, 2),
		sqWithOffset(60, 0, 2),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{0, 0, 0}, geometry.Point{20, 0, 140})
	expectedPath := []geometry.Point{geometry.Point{0, 0}, geometry.Point{60, 0, 60}, geometry.Point{60, 0, 120}, geometry.Point{20, 0, 140}}
	assertPathEq(t, expectedPath, path)
}

func TestC(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(60, 0, 0),
		sqWithOffset(60, 1, 0),
		sqWithOffset(60, 0, 1),
		sqWithOffset(60, 0, 2),
		sqWithOffset(60, 1, 2),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{80, 0, 20}, geometry.Point{80, 0, 140})
	expectedPath := []geometry.Point{geometry.Point{80, 0, 20}, geometry.Point{60, 0, 60}, geometry.Point{60, 0, 120}, geometry.Point{80, 0, 140}}
	assertPathEq(t, expectedPath, path)
}

func TestOnEdgeToApex(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(60, 0, 0),
		sqWithOffset(60, 0, 1),
		sqWithOffset(60, -1, 1),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{0, 0, 30}, geometry.Point{-20, 0, 60})
	expectedPath := []geometry.Point{geometry.Point{0, 0, 30}, geometry.Point{0, 0, 60}, geometry.Point{-20, 0, 60}}
	assertPathEq(t, expectedPath, path)
}

func TestPathDoesNotExist(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(60, 0, 0),
		sqWithOffset(60, 0, 1),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{0, 0, 30}, geometry.Point{61, 0, 0})
	assertPathEq(t, nil, path)
}

func TestStartEqualsGoal(t *testing.T) {
	polygons := []*geometry.Polygon{
		sqWithOffset(60, 0, 0),
	}

	navmesh := ConstructNavMesh(polygons)
	p := Planner{}
	p.SetNavMesh(navmesh)

	path := p.FindPath(geometry.Point{0, 0, 30}, geometry.Point{0, 0, 30})
	expectedPath := []geometry.Point{geometry.Point{0, 0, 30}, geometry.Point{0, 0, 30}}
	assertPathEq(t, expectedPath, path)
}

func sqWithOffset(size, xOffset, yOffset float64) *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{xOffset * size, 0, yOffset * size},
		geometry.Point{xOffset * size, 0, yOffset*size + size},
		geometry.Point{xOffset*size + size, 0, yOffset*size + size},
		geometry.Point{xOffset*size + size, 0, yOffset * size},
	}
	return geometry.NewPolygon(points)
}

func sqWithXOffset(offset float64) *geometry.Polygon {
	points := []geometry.Point{
		geometry.Point{offset + 0, 0, 0},
		geometry.Point{offset + 0, 0, 6},
		geometry.Point{offset + 6, 0, 6},
		geometry.Point{offset + 6, 0, 0},
	}
	return geometry.NewPolygon(points)
}

func assertPathEq(t *testing.T, expected, actual []geometry.Point) {
	if len(actual) != len(expected) {
		t.Fatalf("Expected: %v Actual: %v", expected, actual)
	}

	for i, point := range actual {
		if point != expected[i] {
			t.Fatalf("Expected: %v Actual: %v", expected, actual)
		}
	}
}
