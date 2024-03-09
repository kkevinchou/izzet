package navmesh_test

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/navmesh"
)

func TestHeightField(t *testing.T) {
	hf := navmesh.NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	count := hf.SpanCount()
	if count != 0 {
		t.Fatalf("count %d != 0", count)
	}

	//
	//
	//
	// X
	hf.AddVoxel(0, 0, 0)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	//
	//
	// X
	// X

	// merge
	hf.AddVoxel(0, 1, 0)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}

	// X
	//
	// X
	// X

	hf.AddVoxel(0, 3, 0)
	count = hf.SpanCount()
	if count != 2 {
		t.Fatalf("count %d != 2", count)
	}

	// X
	// X
	// X
	// X

	hf.AddVoxel(0, 2, 0)
	count = hf.SpanCount()
	if count != 1 {
		t.Fatalf("count %d != 1", count)
	}
}
