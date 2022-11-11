package gizmo

import "github.com/go-gl/mathgl/mgl64"

var R *RotationGizmo

type RotationGizmo struct {
	MotionPivot    mgl64.Vec3
	Active         bool
}
