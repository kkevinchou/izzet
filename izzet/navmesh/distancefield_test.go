package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestDistanceField(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 3x3 voxels on xz plane

	hf.AddVoxel(0, 0, 0)
	hf.AddVoxel(1, 0, 0)
	hf.AddVoxel(2, 0, 0)

	hf.AddVoxel(0, 0, 1)
	hf.AddVoxel(1, 0, 1)
	hf.AddVoxel(2, 0, 1)

	hf.AddVoxel(0, 0, 2)
	hf.AddVoxel(1, 0, 2)
	hf.AddVoxel(2, 0, 2)

	chf := NewCompactHeightField(1, 1, hf)
	BuildDistanceField(chf)

	if chf.maxDistance != 2 {
		t.Fatalf("max dist was %d instead of 2", chf.maxDistance)
	}
}

func TestDistanceFieldAllBorders(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 3x3 voxels on xz plane

	hf.AddVoxel(0, 0, 0)
	hf.AddVoxel(1, 0, 0)
	hf.AddVoxel(2, 0, 0)

	hf.AddVoxel(0, 0, 1)
	hf.AddVoxel(2, 0, 1)

	hf.AddVoxel(0, 0, 2)
	hf.AddVoxel(1, 0, 2)
	hf.AddVoxel(2, 0, 2)

	chf := NewCompactHeightField(1, 1, hf)
	BuildDistanceField(chf)

	if chf.maxDistance != 0 {
		t.Fatalf("max dist was %d instead of 0", chf.maxDistance)
	}
}

func TestDistanceFieldBlur(t *testing.T) {
	hf := NewHeightField(7, 7, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 5x5 voxels on xz plane

	hf.AddVoxel(0, 0, 0)
	hf.AddVoxel(1, 0, 0)
	hf.AddVoxel(2, 0, 0)
	hf.AddVoxel(3, 0, 0)
	hf.AddVoxel(4, 0, 0)
	hf.AddVoxel(5, 0, 0)
	hf.AddVoxel(6, 0, 0)

	hf.AddVoxel(0, 0, 1)
	hf.AddVoxel(1, 0, 1)
	hf.AddVoxel(2, 0, 1)
	hf.AddVoxel(3, 0, 1)
	hf.AddVoxel(4, 0, 1)
	hf.AddVoxel(5, 0, 1)
	hf.AddVoxel(6, 0, 1)

	hf.AddVoxel(0, 0, 2)
	hf.AddVoxel(1, 0, 2)
	hf.AddVoxel(2, 0, 2)
	hf.AddVoxel(3, 0, 2)
	hf.AddVoxel(4, 0, 2)
	hf.AddVoxel(5, 0, 2)
	hf.AddVoxel(6, 0, 2)

	hf.AddVoxel(0, 0, 3)
	hf.AddVoxel(1, 0, 3)
	hf.AddVoxel(2, 0, 3)
	hf.AddVoxel(3, 0, 3)
	hf.AddVoxel(4, 0, 3)
	hf.AddVoxel(5, 0, 3)
	hf.AddVoxel(6, 0, 3)

	hf.AddVoxel(0, 0, 4)
	hf.AddVoxel(1, 0, 4)
	hf.AddVoxel(2, 0, 4)
	hf.AddVoxel(3, 0, 4)
	hf.AddVoxel(4, 0, 4)
	hf.AddVoxel(5, 0, 4)
	hf.AddVoxel(6, 0, 4)

	hf.AddVoxel(0, 0, 5)
	hf.AddVoxel(1, 0, 5)
	hf.AddVoxel(2, 0, 5)
	hf.AddVoxel(3, 0, 5)
	hf.AddVoxel(4, 0, 5)
	hf.AddVoxel(5, 0, 5)
	hf.AddVoxel(6, 0, 5)

	hf.AddVoxel(0, 0, 6)
	hf.AddVoxel(1, 0, 6)
	hf.AddVoxel(2, 0, 6)
	hf.AddVoxel(3, 0, 6)
	hf.AddVoxel(4, 0, 6)
	hf.AddVoxel(5, 0, 6)
	hf.AddVoxel(6, 0, 6)

	chf := NewCompactHeightField(1, 1, hf)
	BuildDistanceField(chf)

	twoCostCount := 0
	threeCostCount := 0
	fourCostCount := 0

	for i := 0; i < len(chf.distances); i++ {
		if chf.distances[i] == 2 {
			twoCostCount++
		} else if chf.distances[i] == 3 {
			threeCostCount++
		} else if chf.distances[i] == 4 {
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
