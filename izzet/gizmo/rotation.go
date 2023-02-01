package gizmo

import "github.com/go-gl/mathgl/mgl64"

var R *RotationGizmo

type Circle struct {
	Position mgl64.Vec3
	Normal   mgl64.Vec3
	Radius   float64
}

type RotationGizmo struct {
	MotionPivot mgl64.Vec2
	Axes        []Circle
	Active      bool
	HoverIndex  int
}
