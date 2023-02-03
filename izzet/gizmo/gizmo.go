package gizmo

import "github.com/go-gl/mathgl/mgl64"

type AxisType string

var AxisTypeX AxisType = "X"
var AxisTypeY AxisType = "Y"
var AxisTypeZ AxisType = "Z"

const (
	ActivationRadius = 10
)

type GizmoMode string

var (
	GizmoModeNone        GizmoMode = "NONE"
	GizmoModeTranslation GizmoMode = "TRANSLATION"
	GizmoModeRotation    GizmoMode = "ROTATION"

	CurrentGizmoMode GizmoMode = GizmoModeNone
)

func init() {
	CurrentGizmoMode = GizmoModeNone
	axes := []mgl64.Vec3{mgl64.Vec3{20, 0, 0}, mgl64.Vec3{0, 20, 0}, mgl64.Vec3{0, 0, 20}}
	T = &TranslationGizmo{Axes: axes, HoverIndex: -1}
	R = &RotationGizmo{
		Axes: []Circle{
			Circle{Normal: mgl64.Vec3{0, 0, 1}, Radius: 25},
			Circle{Normal: mgl64.Vec3{1, 0, 0}, Radius: 25},
			Circle{Normal: mgl64.Vec3{0, 1, 0}, Radius: 25},
		},
		HoverIndex: -1}
}
