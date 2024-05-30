package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func setupCHF() *CompactHeightField {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 5x5 voxels on xz plane

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

	return NewCompactHeightField(1, 1, hf)
}

func TestBuildRegion(t *testing.T) {
	chf := setupCHF()

	BuildDistanceField(chf)
	BoxBlur(chf, chf.distances)

	BuildRegions(chf, 999, 1, 1)
}
