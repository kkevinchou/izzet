package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestCompactHeightField(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	hf.AddVoxel(0, 0, 0)
	hf.AddVoxel(1, 0, 0)

	chf := NewCompactHeightField(1, 0, hf)

	if chf.spans[0].neighbors[2] != 1 {
		t.Fatalf("the first span should point to the second span as a neighbor in the positive x direction")
	}
	if chf.spans[1].neighbors[0] != 0 {
		t.Fatalf("the second span should point to the first span as a neighbor in the negative x direction")
	}
}

func TestClimbableHeight(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	hf.AddVoxel(0, 0, 0)
	hf.AddVoxel(0, 1, 0)

	hf.AddVoxel(1, 0, 0)

	chf := NewCompactHeightField(1, 1, hf)

	// still walkable with step size 1
	if chf.spans[0].neighbors[2] != 1 {
		t.Fatalf("the first span should point to the second span as a neighbor in the positive x direction")
	}
	if chf.spans[1].neighbors[0] != 0 {
		t.Fatalf("the second span should point to the first span as a neighbor in the negative x direction")
	}
}

func TestNotClimbableHeight(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	hf.AddVoxel(0, 0, 0)
	hf.AddVoxel(0, 1, 0)
	hf.AddVoxel(0, 2, 0)

	hf.AddVoxel(1, 0, 0)

	chf := NewCompactHeightField(1, 1, hf)

	// not walkable with step size 1
	if chf.spans[0].neighbors[2] != -1 {
		t.Fatalf("the first span should not reach the second span anymore, climbaleHeight of 1, but the difference is 2")
	}
}

func TestNotWalkableHeight(t *testing.T) {
	hf := NewHeightField(100, 100, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{100, 100, 100})

	hf.AddVoxel(0, 0, 0)

	hf.AddVoxel(1, 0, 0)
	hf.AddVoxel(1, 2, 0)

	chf := NewCompactHeightField(3, 0, hf)

	// not walkable with step size 1
	if chf.spans[0].neighbors[2] != -1 {
		t.Fatalf("span should not be reachable due to walkable height")
	}
}
