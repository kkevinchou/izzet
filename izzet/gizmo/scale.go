package gizmo

import "github.com/go-gl/mathgl/mgl64"

var S *ScaleGizmo

type ScaleGizmo struct {
	ScaleDir    mgl64.Vec3
	MotionPivot mgl64.Vec2
	Active      bool
	Axes        []mgl64.Vec3
	HoverIndex  int
}
