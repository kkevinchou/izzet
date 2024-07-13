package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func setupCHF() *CompactHeightField {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 5x5 voxels on xz plane

	hf.AddVoxel(0, 0, 0, true)
	hf.AddVoxel(1, 0, 0, true)
	hf.AddVoxel(2, 0, 0, true)
	hf.AddVoxel(3, 0, 0, true)
	hf.AddVoxel(4, 0, 0, true)

	hf.AddVoxel(0, 0, 1, true)
	hf.AddVoxel(1, 0, 1, true)
	hf.AddVoxel(2, 0, 1, true)
	hf.AddVoxel(3, 0, 1, true)
	hf.AddVoxel(4, 0, 1, true)

	hf.AddVoxel(0, 0, 2, true)
	hf.AddVoxel(1, 0, 2, true)
	hf.AddVoxel(2, 0, 2, true)
	hf.AddVoxel(3, 0, 2, true)
	hf.AddVoxel(4, 0, 2, true)

	hf.AddVoxel(0, 0, 3, true)
	hf.AddVoxel(1, 0, 3, true)
	hf.AddVoxel(2, 0, 3, true)
	hf.AddVoxel(3, 0, 3, true)
	hf.AddVoxel(4, 0, 3, true)

	hf.AddVoxel(0, 0, 4, true)
	hf.AddVoxel(1, 0, 4, true)
	hf.AddVoxel(2, 0, 4, true)
	hf.AddVoxel(3, 0, 4, true)
	hf.AddVoxel(4, 0, 4, true)

	return NewCompactHeightField(1, 1, hf)
}

func TestBuildRegion(t *testing.T) {
	chf := setupCHF()

	BuildDistanceField(chf)
	BoxBlur(chf, chf.Distances)

	BuildRegions(chf, 999, 1, 1)
}
