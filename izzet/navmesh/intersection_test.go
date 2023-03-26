package navmesh

import (
	"fmt"
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestIntersection(t *testing.T) {
	aabb := AABB{
		Min: mgl64.Vec3{-1, -1, -1},
		Max: mgl64.Vec3{1, 1, 1},
	}

	triangle := Triangle{
		V1: mgl64.Vec3{0.5, 0.5, 2},
		V2: mgl64.Vec3{1, 0.5, 2},
		V3: mgl64.Vec3{1, 1, 2},
	}

	intersection := IntersectAABBTriangle(aabb, triangle)
	fmt.Println(intersection)
	t.Fail()
}
