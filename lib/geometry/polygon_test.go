package geometry

import "testing"

func defaultPolygon() *Polygon {
	points := []Point{
		Point{0, 0, 0},
		Point{0, 0, 6},
		Point{6, 0, 6},
		Point{6, 0, 0},
	}
	return NewPolygon(points)
}

func assert(t *testing.T, errorMessage string, actual, expected bool) {
	if expected != actual {
		t.Fatalf("%s - Expected [%v] but got [%v]", errorMessage, expected, actual)
	}
}

// We consider the borders to be inclusive, may be subject to change in the future
func TestContainsPointOnBorder(t *testing.T) {
	polygon := defaultPolygon()

	assert(t, "Should contain point when it lies on the top border", polygon.ContainsPoint(Point{3, 0, 0}), true)
	assert(t, "Should contain point when it lies on the bottom border", polygon.ContainsPoint(Point{3, 0, 6}), true)
	assert(t, "Should contain point when it lies on the left border", polygon.ContainsPoint(Point{0, 0, 3}), true)
	assert(t, "Should contain point when it lies on the right border", polygon.ContainsPoint(Point{0, 0, 3}), true)

	assert(t, "Should contain point when it overlaps a point", polygon.ContainsPoint(Point{0, 0, 0}), true)
	assert(t, "Should contain point when it overlaps a point", polygon.ContainsPoint(Point{0, 0, 6}), true)
	assert(t, "Should contain point when it overlaps a point", polygon.ContainsPoint(Point{6, 0, 6}), true)
	assert(t, "Should contain point when it overlaps a point", polygon.ContainsPoint(Point{6, 0, 0}), true)
}

func TestContainsPointWithinBorder(t *testing.T) {
	polygon := defaultPolygon()
	assert(t, "Should contain point when it lies within the borders", polygon.ContainsPoint(Point{3, 0, 3}), true)
	assert(t, "Should contain point when it lies within the borders", polygon.ContainsPoint(Point{1, 0, 2}), true)
	assert(t, "Should contain point when it lies within the borders", polygon.ContainsPoint(Point{5, 0, 4}), true)
	assert(t, "Should contain point when it lies within the borders", polygon.ContainsPoint(Point{3, 0, 1}), true)
}

func TestDoesNotContainPoint(t *testing.T) {
	polygon := defaultPolygon()
	assert(t, "Should contain point when it lies above the polygon", polygon.ContainsPoint(Point{3, 0, -10}), false)
	assert(t, "Should contain point when it lies below the polygon", polygon.ContainsPoint(Point{3, 0, 10}), false)
	assert(t, "Should contain point when it lies left of the polygon", polygon.ContainsPoint(Point{10, 0, 3}), false)
	assert(t, "Should contain point when it lies right of the polygon", polygon.ContainsPoint(Point{-10, 0, 3}), false)
}
