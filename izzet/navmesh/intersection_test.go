package navmesh

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

type Case struct {
	aabb           AABB
	triangle       Triangle
	expectedResult bool
}

var cases []Case = []Case{
	{
		aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
		triangle:       Triangle{V1: mgl64.Vec3{0.5, 0.5, 2}, V2: mgl64.Vec3{1, 0.5, 2}, V3: mgl64.Vec3{1, 1, 2}},
		expectedResult: false,
	},
	{
		aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
		triangle:       Triangle{V1: mgl64.Vec3{0.5, 0.5, 1}, V2: mgl64.Vec3{1, 0.5, 1}, V3: mgl64.Vec3{1, 1, 1}},
		expectedResult: true,
	},
	{
		aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
		triangle:       Triangle{V1: mgl64.Vec3{-0.5, 0.5, 1}, V2: mgl64.Vec3{-1, 0.5, 1}, V3: mgl64.Vec3{-1, 1, 1}},
		expectedResult: true,
	},
	{
		aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
		triangle:       Triangle{V1: mgl64.Vec3{0, 0, 0}, V2: mgl64.Vec3{0, 0, 1}, V3: mgl64.Vec3{1.5, 0, 1}},
		expectedResult: true,
	},
	{
		aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
		triangle:       Triangle{V1: mgl64.Vec3{0, 0, 0}, V2: mgl64.Vec3{0, 0, 1}, V3: mgl64.Vec3{1.5, 0, 1}},
		expectedResult: true,
	},
	{
		aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
		triangle:       Triangle{V1: mgl64.Vec3{-0.5, 1.5, 1}, V2: mgl64.Vec3{-1, 1.5, 1}, V3: mgl64.Vec3{-1, 1.5, 1}},
		expectedResult: false,
	},
}

func TestIntersection(t *testing.T) {
	for _, c := range cases {
		intersection := IntersectAABBTriangle(c.aabb, c.triangle)

		if intersection != c.expectedResult {
			t.Fail()
		}
	}
}
