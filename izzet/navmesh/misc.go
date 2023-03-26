package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/collision/collider"
)

func bbVerts(bb collider.BoundingBox) []mgl64.Vec3 {
	min := bb.MinVertex
	max := bb.MaxVertex
	delta := max.Sub(min)

	verts := []mgl64.Vec3{
		// top
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		max,
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),

		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),
		max,

		// bottom
		min,
		min.Add(mgl64.Vec3{delta[0], 0, 0}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),

		min,
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),
		min.Add(mgl64.Vec3{0, 0, delta[2]}),

		// left
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min,
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		min,
		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		// right
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, 0}),

		min.Add(mgl64.Vec3{delta[0], 0, 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),

		// front
		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),

		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		// back
		min,
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], 0, 0}),

		min,
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
	}
	return verts
}
