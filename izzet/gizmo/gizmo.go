package gizmo

import "github.com/go-gl/mathgl/mgl64"

const (
	ActivationRadius = 10
)

type GizmoMode string

var (
	GizmoModeNone        GizmoMode = "NONE"
	GizmoModeTranslation GizmoMode = "TRANSLATION"
	GizmoModeRotation    GizmoMode = "ROTATION"
	GizmoModeScale       GizmoMode = "SCALE"

	CurrentGizmoMode GizmoMode = GizmoModeNone
)

var (
	GizmoXAxis mgl64.Vec3 = mgl64.Vec3{1, 0, 0}
	GizmoYAxis mgl64.Vec3 = mgl64.Vec3{0, 1, 0}
	GizmoZAxis mgl64.Vec3 = mgl64.Vec3{0, 0, 1}
)

func init() {
	CurrentGizmoMode = GizmoModeNone
	axes := []mgl64.Vec3{GizmoXAxis, GizmoYAxis, GizmoZAxis}
	T = &TranslationGizmo{Axes: axes, HoverIndex: -1}

	R = &RotationGizmo{
		Axes: []Circle{
			Circle{Normal: mgl64.Vec3{0, 0, 1}, Radius: 25},
			Circle{Normal: mgl64.Vec3{1, 0, 0}, Radius: 25},
			Circle{Normal: mgl64.Vec3{0, 1, 0}, Radius: 25},
		},
		HoverIndex: -1}

	segments := []Axis{
		Axis{Vector: GizmoXAxis, Type: XAxis},
		Axis{Vector: GizmoYAxis, Type: YAxis},
		Axis{Vector: GizmoZAxis, Type: ZAxis},
	}
	S = &ScaleGizmo{Axes: segments, HoveredAxisType: NullAxis}
}
