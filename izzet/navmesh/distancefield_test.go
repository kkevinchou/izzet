package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestDistanceField(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 3x3 voxels on xz plane

	hf.AddSpan(0, 0, 0, 1, true, 1)
	hf.AddSpan(1, 0, 0, 1, true, 1)
	hf.AddSpan(2, 0, 0, 1, true, 1)

	hf.AddSpan(0, 1, 0, 1, true, 1)
	hf.AddSpan(1, 1, 0, 1, true, 1)
	hf.AddSpan(2, 1, 0, 1, true, 1)

	hf.AddSpan(0, 2, 0, 1, true, 1)
	hf.AddSpan(1, 2, 0, 1, true, 1)
	hf.AddSpan(2, 2, 0, 1, true, 1)

	chf := NewCompactHeightField(1, 1, hf)
	BuildDistanceField(chf)

	if chf.maxDistance != 2 {
		t.Fatalf("max dist was %d instead of 2", chf.maxDistance)
	}
}

func TestDistanceFieldAllBorders(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 3x3 voxels on xz plane

	hf.AddSpan(0, 0, 0, 1, true, 1)
	hf.AddSpan(1, 0, 0, 1, true, 1)
	hf.AddSpan(2, 0, 0, 1, true, 1)

	hf.AddSpan(0, 1, 0, 1, true, 1)
	hf.AddSpan(2, 1, 0, 1, true, 1)

	hf.AddSpan(0, 2, 0, 1, true, 1)
	hf.AddSpan(1, 2, 0, 1, true, 1)
	hf.AddSpan(2, 2, 0, 1, true, 1)

	chf := NewCompactHeightField(1, 1, hf)
	BuildDistanceField(chf)

	if chf.maxDistance != 0 {
		t.Fatalf("max dist was %d instead of 0", chf.maxDistance)
	}
}

func TestDistanceFieldBlur(t *testing.T) {
	hf := NewHeightField(7, 7, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 5x5 voxels on xz plane

	hf.AddSpan(0, 0, 0, 1, true, 1)
	hf.AddSpan(1, 0, 0, 1, true, 1)
	hf.AddSpan(2, 0, 0, 1, true, 1)
	hf.AddSpan(3, 0, 0, 1, true, 1)
	hf.AddSpan(4, 0, 0, 1, true, 1)
	hf.AddSpan(5, 0, 0, 1, true, 1)
	hf.AddSpan(6, 0, 0, 1, true, 1)

	hf.AddSpan(0, 1, 0, 1, true, 1)
	hf.AddSpan(1, 1, 0, 1, true, 1)
	hf.AddSpan(2, 1, 0, 1, true, 1)
	hf.AddSpan(3, 1, 0, 1, true, 1)
	hf.AddSpan(4, 1, 0, 1, true, 1)
	hf.AddSpan(5, 1, 0, 1, true, 1)
	hf.AddSpan(6, 1, 0, 1, true, 1)

	hf.AddSpan(0, 2, 0, 1, true, 1)
	hf.AddSpan(1, 2, 0, 1, true, 1)
	hf.AddSpan(2, 2, 0, 1, true, 1)
	hf.AddSpan(3, 2, 0, 1, true, 1)
	hf.AddSpan(4, 2, 0, 1, true, 1)
	hf.AddSpan(5, 2, 0, 1, true, 1)
	hf.AddSpan(6, 2, 0, 1, true, 1)

	hf.AddSpan(0, 3, 0, 1, true, 1)
	hf.AddSpan(1, 3, 0, 1, true, 1)
	hf.AddSpan(2, 3, 0, 1, true, 1)
	hf.AddSpan(3, 3, 0, 1, true, 1)
	hf.AddSpan(4, 3, 0, 1, true, 1)
	hf.AddSpan(5, 3, 0, 1, true, 1)
	hf.AddSpan(6, 3, 0, 1, true, 1)

	hf.AddSpan(0, 4, 0, 1, true, 1)
	hf.AddSpan(1, 4, 0, 1, true, 1)
	hf.AddSpan(2, 4, 0, 1, true, 1)
	hf.AddSpan(3, 4, 0, 1, true, 1)
	hf.AddSpan(4, 4, 0, 1, true, 1)
	hf.AddSpan(5, 4, 0, 1, true, 1)
	hf.AddSpan(6, 4, 0, 1, true, 1)

	hf.AddSpan(0, 5, 0, 1, true, 1)
	hf.AddSpan(1, 5, 0, 1, true, 1)
	hf.AddSpan(2, 5, 0, 1, true, 1)
	hf.AddSpan(3, 5, 0, 1, true, 1)
	hf.AddSpan(4, 5, 0, 1, true, 1)
	hf.AddSpan(5, 5, 0, 1, true, 1)
	hf.AddSpan(6, 5, 0, 1, true, 1)

	hf.AddSpan(0, 6, 0, 1, true, 1)
	hf.AddSpan(1, 6, 0, 1, true, 1)
	hf.AddSpan(2, 6, 0, 1, true, 1)
	hf.AddSpan(3, 6, 0, 1, true, 1)
	hf.AddSpan(4, 6, 0, 1, true, 1)
	hf.AddSpan(5, 6, 0, 1, true, 1)
	hf.AddSpan(6, 6, 0, 1, true, 1)

	chf := NewCompactHeightField(1, 1, hf)
	BuildDistanceField(chf)

	twoCostCount := 0
	threeCostCount := 0
	fourCostCount := 0

	for i := 0; i < len(chf.Distances); i++ {
		if chf.Distances[i] == 2 {
			twoCostCount++
		} else if chf.Distances[i] == 3 {
			threeCostCount++
		} else if chf.Distances[i] == 4 {
			fourCostCount++
		}
	}

	if twoCostCount != 16 {
		t.Fatalf("two cost count should be 16, got %d", twoCostCount)
	}

	if threeCostCount != 4 {
		t.Fatalf("three cost count should be 4, got %d", threeCostCount)
	}

	if fourCostCount != 5 {
		t.Fatalf("four cost count should be 5, got %d", fourCostCount)
	}
}
