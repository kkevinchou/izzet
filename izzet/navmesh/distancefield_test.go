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

	_, maxDist := BuildDistanceField(chf)

	if maxDist != 2 {
		t.Fatalf("max dist was %d instead of 2", maxDist)
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

	_, maxDist := BuildDistanceField(chf)

	if maxDist != 0 {
		t.Fatalf("max dist was %d instead of 0", maxDist)
	}
}

func TestDistanceFieldBlur(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 3x3 voxels on xz plane

	hf.AddVoxel(0, 0, 0)
	hf.AddVoxel(1, 0, 0)
	hf.AddVoxel(2, 0, 0)
	hf.AddVoxel(3, 0, 0)
	hf.AddVoxel(4, 0, 0)

	hf.AddVoxel(0, 0, 1)
	hf.AddVoxel(1, 0, 1)
	hf.AddVoxel(2, 0, 1)
	hf.AddVoxel(3, 0, 1)
	hf.AddVoxel(4, 0, 1)

	hf.AddVoxel(0, 0, 2)
	hf.AddVoxel(1, 0, 2)
	hf.AddVoxel(2, 0, 2)
	hf.AddVoxel(3, 0, 2)
	hf.AddVoxel(4, 0, 2)

	hf.AddVoxel(0, 0, 3)
	hf.AddVoxel(1, 0, 3)
	hf.AddVoxel(2, 0, 3)
	hf.AddVoxel(3, 0, 3)
	hf.AddVoxel(4, 0, 3)

	hf.AddVoxel(0, 0, 4)
	hf.AddVoxel(1, 0, 4)
	hf.AddVoxel(2, 0, 4)
	hf.AddVoxel(3, 0, 4)
	hf.AddVoxel(4, 0, 4)

	chf := NewCompactHeightField(1, 1, hf)
	distances, _ := BuildDistanceField(chf)
	blurredDistances := BoxBlur(chf, distances)

	twoCostCount := 0
	threeCostCount := 0

	for i := 0; i < len(blurredDistances); i++ {
		if blurredDistances[i] == 2 {
			twoCostCount++
		} else if blurredDistances[i] == 3 {
			threeCostCount++
		}
	}

	if twoCostCount != 8 {
		t.Fatalf("two cost count should be 8, got %d", twoCostCount)
	}

	if threeCostCount != 1 {
		t.Fatalf("three cost count should be 1, got %d", threeCostCount)
	}
}
