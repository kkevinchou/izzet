package gizmo

import "github.com/go-gl/mathgl/mgl64"

var R *RotationGizmo

type Circle struct {
	Normal mgl64.Vec3
	Radius float64
}

type RotationGizmo struct {
	MotionPivot        mgl64.Vec2
	Axes               []Circle
	Active             bool
	HoverIndex         int
	ActivationRotation mgl64.Quat
}

func (g *RotationGizmo) Reset() {
	g.HoverIndex = -1
	g.ActivationRotation = mgl64.QuatIdent()
	g.Active = false
}
