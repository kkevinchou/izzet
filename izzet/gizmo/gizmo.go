package gizmo

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/constants"
)

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

	TGizmo *Gizmo
	SGizmo *Gizmo
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

	TGizmo = setupTranslationGizmo()
	SGizmo = setupTranslationGizmo()
}

type GizmoAxis struct {
	Direction mgl64.Vec3
}

type Gizmo struct {
	EntityIDToAxis        map[int]GizmoAxis
	HoveredEntityID       int
	Active                bool
	LastFrameClosestPoint mgl64.Vec3
}

func setupTranslationGizmo() *Gizmo {
	return &Gizmo{
		HoveredEntityID: -1,
		EntityIDToAxis: map[int]GizmoAxis{
			constants.GizmoXAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{1, 0, 0}},
			constants.GizmoYAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{0, 1, 0}},
			constants.GizmoZAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{0, 0, 1}},
		},
	}
}

func setupScaleGizmo() *Gizmo {
	return &Gizmo{
		HoveredEntityID: -1,
		EntityIDToAxis: map[int]GizmoAxis{
			constants.GizmoXAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{1, 0, 0}},
			constants.GizmoYAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{0, 1, 0}},
			constants.GizmoZAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{0, 0, 1}},
		},
	}
}

func (g *Gizmo) Reset() {
	g.HoveredEntityID = -1
	g.Active = false
}
