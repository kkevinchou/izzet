package gizmo

import (
	"github.com/go-gl/mathgl/mgl64"
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
	TranslationGizmo *Gizmo
	ScaleGizmo       *Gizmo
)

const (
	GizmoXAxisPickingID   int = 1000000000
	GizmoYAxisPickingID   int = 1000000001
	GizmoZAxisPickingID   int = 1000000002
	GizmoAllAxisPickingID int = 1000000003
)

func init() {
	CurrentGizmoMode = GizmoModeNone

	R = &RotationGizmo{
		Axes: []Circle{
			Circle{Normal: mgl64.Vec3{0, 0, 1}, Radius: 25},
			Circle{Normal: mgl64.Vec3{1, 0, 0}, Radius: 25},
			Circle{Normal: mgl64.Vec3{0, 1, 0}, Radius: 25},
		},
		HoverIndex: -1}

	TranslationGizmo = setupTranslationGizmo()
	ScaleGizmo = setupScaleGizmo()
}

type GizmoAxis struct {
	Direction mgl64.Vec3
}

type Gizmo struct {
	EntityIDToAxis         map[int]GizmoAxis
	HoveredEntityID        int
	Active                 bool
	LastFrameClosestPoint  mgl64.Vec3
	LastFrameMousePosition mgl64.Vec2
}

func setupTranslationGizmo() *Gizmo {
	return &Gizmo{
		HoveredEntityID: -1,
		EntityIDToAxis: map[int]GizmoAxis{
			GizmoXAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{1, 0, 0}},
			GizmoYAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{0, 1, 0}},
			GizmoZAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{0, 0, 1}},
		},
	}
}

func setupScaleGizmo() *Gizmo {
	return &Gizmo{
		HoveredEntityID: -1,
		EntityIDToAxis: map[int]GizmoAxis{
			GizmoXAxisPickingID:   GizmoAxis{Direction: mgl64.Vec3{1, 0, 0}},
			GizmoYAxisPickingID:   GizmoAxis{Direction: mgl64.Vec3{0, 1, 0}},
			GizmoZAxisPickingID:   GizmoAxis{Direction: mgl64.Vec3{0, 0, 1}},
			GizmoAllAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{}},
		},
	}
}

func (g *Gizmo) Reset() {
	g.HoveredEntityID = -1
	g.Active = false
}
