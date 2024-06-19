package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestHeightField(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	count := hf.SpanCount()
	if count != 0 {
		t.Fatalf("count %d != 0", count)
	}

	//
	//
	//
	// X

	// first voxel
	hf.AddVoxel(0, 0, 0, true)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	//
	//
	// X
	// X

	// merge
	hf.AddVoxel(0, 1, 0, true)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	// X
	//
	// X
	// X

	// no merge
	hf.AddVoxel(0, 3, 0, true)
	count = hf.SpanCount()
	if count != 2 {
		t.Fatalf("count %d != 2", count)
	}

	// X
	// X
	// X
	// X

	// merge it all
	hf.AddVoxel(0, 2, 0, true)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	// X
	// X
	// X
	// X

	// nothing happens at the top
	hf.AddVoxel(0, 3, 0, true)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	// nothing happens at the bottom
	hf.AddVoxel(0, 0, 0, true)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}
}

func TestFilterLowHeightSpans(t *testing.T) {
	walkableHeight := 5
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})
	hf.AddVoxel(0, 0, 0, true)

	// okay
	hf.AddVoxel(0, 5, 0, true)
	FilterLowHeightSpans(walkableHeight, hf)

	if hf.spans[0].area == NULL_AREA {
		t.Fatalf("span should be valid")
	}

	// not okay
	hf.AddVoxel(0, 4, 0, true)
	FilterLowHeightSpans(5, hf)

	if hf.spans[0].area == WALKABLE_AREA {
		t.Fatalf("span should be invalid")
	}
}
