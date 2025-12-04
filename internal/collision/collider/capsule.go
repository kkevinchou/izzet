package collider

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/utils"
)

// A capsule is defined by a line segment with a radius around all points
// along the line segment.
// Top and Bottom represents the top and bottom points of the line segment
type Capsule struct {
	Radius float64
	Top    mgl64.Vec3
	Bottom mgl64.Vec3
}

func NewCapsule(top, bottom mgl64.Vec3, radius float64) Capsule {
	return Capsule{
		Radius: radius,
		Top:    top,
		Bottom: bottom,
	}
}

func (c Capsule) Transform(transform mgl64.Mat4) Capsule {
	top := transform.Mul4x1(c.Top.Vec4(1)).Vec3()
	bottom := transform.Mul4x1(c.Bottom.Vec4(1)).Vec3()
	_, _, scaleVec := utils.DecomposeF64(transform)

	scale := math.Max(scaleVec[0], math.Max(scaleVec[1], scaleVec[2]))

	// assume universal scale
	return NewCapsule(top, bottom, c.Radius*scale)
}

// NewCapsuleFromVertices creates a capsule centered at the model space origin
func NewCapsuleFromVertices(vertices []mgl64.Vec3) Capsule {
	var minX, minY, minZ, maxX, maxY, maxZ float64

	minX = vertices[0].X()
	maxX = vertices[0].X()
	minY = vertices[0].Y()
	maxY = vertices[0].Y()
	minZ = vertices[0].Z()
	maxZ = vertices[0].Z()
	for _, vertex := range vertices {
		if vertex.X() < minX {
			minX = vertex.X()
		}
		if vertex.X() > maxX {
			maxX = vertex.X()
		}
		if vertex.Y() < minY {
			minY = vertex.Y()
		}
		if vertex.Y() > maxY {
			maxY = vertex.Y()
		}
		if vertex.Z() < minZ {
			minZ = vertex.Z()
		}
		if vertex.Z() > maxZ {
			maxZ = vertex.Z()
		}
	}

	radius := math.Abs(minX)
	radius = math.Max(radius, math.Abs(maxX))
	radius = math.Max(radius, math.Abs(minZ))
	radius = math.Max(radius, math.Abs(maxZ))

	// t-pose usually makes the model wider than it needs to be so we shrink radius by a factor
	radius /= 2

	// try our best to construct a capsule that sits above ground
	topYValue := math.Max((2*radius)+1, maxY)

	return Capsule{
		Radius: radius,
		Bottom: mgl64.Vec3{0, radius, 0},
		Top:    mgl64.Vec3{0, topYValue - radius, 0},
	}
}
