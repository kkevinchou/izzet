package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func setupCHF() *CompactHeightField {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	// 5x5 voxels on xz plane

	hf.AddSpan(0, 0, 0, 1, true, 1)
	hf.AddSpan(1, 0, 0, 1, true, 1)
	hf.AddSpan(2, 0, 0, 1, true, 1)
	hf.AddSpan(3, 0, 0, 1, true, 1)
	hf.AddSpan(4, 0, 0, 1, true, 1)

	hf.AddSpan(0, 1, 0, 1, true, 1)
	hf.AddSpan(1, 1, 0, 1, true, 1)
	hf.AddSpan(2, 1, 0, 1, true, 1)
	hf.AddSpan(3, 1, 0, 1, true, 1)
	hf.AddSpan(4, 1, 0, 1, true, 1)

	hf.AddSpan(0, 2, 0, 1, true, 1)
	hf.AddSpan(1, 2, 0, 1, true, 1)
	hf.AddSpan(2, 2, 0, 1, true, 1)
	hf.AddSpan(3, 2, 0, 1, true, 1)
	hf.AddSpan(4, 2, 0, 1, true, 1)

	hf.AddSpan(0, 3, 0, 1, true, 1)
	hf.AddSpan(1, 3, 0, 1, true, 1)
	hf.AddSpan(2, 3, 0, 1, true, 1)
	hf.AddSpan(3, 3, 0, 1, true, 1)
	hf.AddSpan(4, 3, 0, 1, true, 1)

	hf.AddSpan(0, 4, 0, 1, true, 1)
	hf.AddSpan(1, 4, 0, 1, true, 1)
	hf.AddSpan(2, 4, 0, 1, true, 1)
	hf.AddSpan(3, 4, 0, 1, true, 1)
	hf.AddSpan(4, 4, 0, 1, true, 1)

	return NewCompactHeightField(1, 1, hf)
}

func TestBuildRegion(t *testing.T) {
	chf := setupCHF()

	BuildDistanceField(chf)
	BoxBlur(chf, chf.Distances)

	BuildRegions(chf, 999, 1, 1)
}
