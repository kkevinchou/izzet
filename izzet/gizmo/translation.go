package gizmo

import "github.com/go-gl/mathgl/mgl64"

var T *TranslationGizmo

type AxisType string

var AxisTypeX AxisType = "X"
var AxisTypeY AxisType = "Y"
var AxisTypeZ AxisType = "Z"

const (
	ActivationRadius = 10
)

func init() {
	axes := []mgl64.Vec3{mgl64.Vec3{20, 0, 0}, mgl64.Vec3{0, 20, 0}, mgl64.Vec3{0, 0, 20}}
	T = &TranslationGizmo{Axes: axes}
}

type TranslationGizmo struct {
	TranslationDir mgl64.Vec3
	MotionPivot    mgl64.Vec3
	Active         bool
	Axes           []mgl64.Vec3
	HoverIndex     int
}
