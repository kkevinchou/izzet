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
