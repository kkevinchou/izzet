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
	hf.AddSpan(0, 0, 0, 1, true, 1)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	//
	//
	// X
	// X

	// merge
	hf.AddSpan(0, 0, 1, 2, true, 1)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	// X
	//
	// X
	// X

	// no merge
	hf.AddSpan(0, 0, 3, 4, true, 1)
	count = hf.SpanCount()
	if count != 2 {
		t.Fatalf("count %d != 2", count)
	}

	// X
	// X
	// X
	// X

	// merge it all
	hf.AddSpan(0, 0, 2, 3, true, 1)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	// X
	// X
	// X
	// X

	// nothing happens at the top
	hf.AddSpan(0, 0, 3, 4, true, 1)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	// nothing happens at the bottom
	hf.AddSpan(0, 0, 0, 1, true, 1)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}
}

func TestFilterLowHeightSpans(t *testing.T) {
	walkableHeight := 4
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})
	hf.AddSpan(0, 0, 0, 1, true, 1)

	// okay
	hf.AddSpan(0, 0, 5, 6, true, 1)
	FilterLowHeightSpans(walkableHeight, hf)

	if hf.Spans[0].Area == NULL_AREA {
		t.Fatalf("span should be valid")
	}

	// not okay
	hf.AddSpan(0, 0, 4, 5, true, 1)
	FilterLowHeightSpans(5, hf)

	if hf.Spans[0].Area == WALKABLE_AREA {
		t.Fatalf("span should be invalid")
	}
}
