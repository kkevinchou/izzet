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
	// {
	// 	aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
	// 	triangle:       Triangle{V1: mgl64.Vec3{0.5, 0.5, 2}, V2: mgl64.Vec3{1, 0.5, 2}, V3: mgl64.Vec3{1, 1, 2}},
	// 	expectedResult: false,
	// },
	// {
	// 	aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
	// 	triangle:       Triangle{V1: mgl64.Vec3{0.5, 0.5, 1}, V2: mgl64.Vec3{1, 0.5, 1}, V3: mgl64.Vec3{1, 1, 1}},
	// 	expectedResult: true,
	// },
	// {
	// 	aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
	// 	triangle:       Triangle{V1: mgl64.Vec3{-0.5, 0.5, 1}, V2: mgl64.Vec3{-1, 0.5, 1}, V3: mgl64.Vec3{-1, 1, 1}},
	// 	expectedResult: true,
	// },
	// {
	// 	aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
	// 	triangle:       Triangle{V1: mgl64.Vec3{0, 0, 0}, V2: mgl64.Vec3{0, 0, 1}, V3: mgl64.Vec3{1.5, 0, 1}},
	// 	expectedResult: true,
	// },
	// {
	// 	aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
	// 	triangle:       Triangle{V1: mgl64.Vec3{0, 0, 0}, V2: mgl64.Vec3{0, 0, 1}, V3: mgl64.Vec3{1.5, 0, 1}},
	// 	expectedResult: true,
	// },
	// {
	// 	aabb:           AABB{Min: mgl64.Vec3{-1, -1, -1}, Max: mgl64.Vec3{1, 1, 1}},
	// 	triangle:       Triangle{V1: mgl64.Vec3{-0.5, 1.5, 1}, V2: mgl64.Vec3{-1, 1.5, 1}, V3: mgl64.Vec3{-1, 1.5, 1}},
	// 	expectedResult: false,
	// },
	// {
	// 	aabb: AABB{Min: mgl64.Vec3{-101, -1, -90}, Max: mgl64.Vec3{-100, 0, -89}},
	// 	triangle: Triangle{
	// 		V1: mgl64.Vec3{0.47112274169921875, 0.8513708710670471, 2.376190185546875},
	// 		V2: mgl64.Vec3{0.46967315673828125, 0.7559960782527924, -4.459754943847656},
	// 		V3: mgl64.Vec3{0.5, 0.5000000000000007, 0.22772979736328125},
	// 	},
	// 	expectedResult: true,
	// },
	{
		aabb: AABB{Min: mgl64.Vec3{0, 0, 100}, Max: mgl64.Vec3{1, 1, 101}},
		triangle: Triangle{
			// V1: mgl64.Vec3{0.019916534423828125, 0.4266499876976013, 102.72208404541016},
			// V2: mgl64.Vec3{0.5111160278320312, 0.3816499710083008, 102.5407943725586},
			// V3: mgl64.Vec3{3.814697265625e-06, 3.0616166814766664e-15, 99.99999237060547},
			V1: mgl64.Vec3{0.019916534423828125, 0.4266499876976013, 102.72208404541016},
			V2: mgl64.Vec3{0.5111160278320312, 0.3816499710083008, 102.5407943725586},
			V3: mgl64.Vec3{3.814697265625e-06, 3.0616166814766664e-15, 99.99999237060547},
		},
		expectedResult: true,
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
