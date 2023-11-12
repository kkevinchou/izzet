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
	RotationGizmo    *Gizmo
	ScaleGizmo       *Gizmo
)

const (
	GizmoXAxisPickingID   int = 1000000000
	GizmoYAxisPickingID   int = 1000000001
	GizmoZAxisPickingID   int = 1000000002
	GizmoAllAxisPickingID int = 1000000003

	GizmoXDistancePickingID int = 1000000004
	GizmoYDistancePickingID int = 1000000005
	GizmoZDistancePickingID int = 1000000006
)

func init() {
	CurrentGizmoMode = GizmoModeNone

	TranslationGizmo = setupTranslationGizmo()
	RotationGizmo = setupRotationGizmo()
	ScaleGizmo = setupScaleGizmo()
}

type GizmoAxis struct {
	DistanceBasedDelta bool
	Direction          mgl64.Vec3
}

type Gizmo struct {
	EntityIDToAxis         map[int]GizmoAxis
	HoveredEntityID        int
	Active                 bool
	LastFrameClosestPoint  mgl64.Vec3
	LastFrameMousePosition mgl64.Vec2

	ActivationPosition mgl64.Vec3
	ActivationScale    mgl64.Vec3
	ActivationRotation mgl64.Quat
}

// func calculateGizmoDelta(targetGizmo *Gizmo, frameInput input.Input, gizmoPosition mgl64.Vec3, cameraViewDirection mgl64.Vec3, cameraPosition mgl64.Vec3) *mgl64.Vec3 {
// 	// startStatus := targetGizmo.Active
// 	mouseInput := frameInput.MouseInput

// 	colorPickingID := g.renderer.GetEntityByPixelPosition(mouseInput.Position, g.height)
// 	if colorPickingID != nil {
// 		if _, ok := targetGizmo.EntityIDToAxis[*colorPickingID]; ok {
// 			if !mouseInput.Buttons[0] {
// 				targetGizmo.HoveredEntityID = *colorPickingID
// 			}
// 		} else {
// 			colorPickingID = nil
// 		}
// 	}

// 	if colorPickingID != nil {
// 		if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
// 			axis := targetGizmo.EntityIDToAxis[*colorPickingID]
// 			if axis.DistanceBasedDelta {
// 				targetGizmo.LastFrameMousePosition = mouseInput.Position
// 			} else if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(cameraPosition, cameraPosition.Add(cameraViewDirection), gizmoPosition, gizmoPosition.Add(axis.Direction)); nonParallel {
// 				targetGizmo.LastFrameClosestPoint = closestPointOnAxis
// 				targetGizmo.LastFrameMousePosition = mouseInput.Position
// 			} else if !nonParallel && *colorPickingID == GizmoAllAxisPickingID {
// 				targetGizmo.LastFrameClosestPoint = closestPointOnAxis
// 				targetGizmo.LastFrameMousePosition = mouseInput.Position
// 			} else {
// 				panic("parallel")
// 			}

// 			targetGizmo.Active = true
// 		}
// 	} else if !targetGizmo.Active {
// 		// specifically check that the gizmo is not active before reseting.
// 		// this supports the scenario where we initially click and drag a gizmo
// 		// to the point where the mouse leaves the range of any axes
// 		targetGizmo.Reset()
// 	}

// 	if !targetGizmo.Active {
// 		return nil
// 	}

// 	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
// 		targetGizmo.Reset()
// 	}

// 	var gizmoDelta *mgl64.Vec3

// 	if mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
// 		axis := targetGizmo.EntityIDToAxis[targetGizmo.HoveredEntityID]

// 		if axis.DistanceBasedDelta {
// 			// mouse position based deltas, store the x,y mouse delta in the return value with 0 for the z value
// 			mouseDelta := mouseInput.Position.Sub(targetGizmo.LastFrameMousePosition).Vec3(0)
// 			gizmoDelta = &mouseDelta
// 			targetGizmo.LastFrameMousePosition = mouseInput.Position
// 		} else if targetGizmo.HoveredEntityID == GizmoAllAxisPickingID {
// 			mouseDelta := mouseInput.Position.Sub(targetGizmo.LastFrameMousePosition)
// 			magnitude := (mouseDelta[0] - mouseDelta[1])
// 			delta := mgl64.Vec3{1, 1, 1}.Mul(magnitude)
// 			gizmoDelta = &delta
// 			targetGizmo.LastFrameMousePosition = mouseInput.Position
// 		} else {
// 			if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(cameraPosition, cameraPosition.Add(cameraViewDirection), gizmoPosition, gizmoPosition.Add(axis.Direction)); nonParallel {
// 				delta := closestPointOnAxis.Sub(targetGizmo.LastFrameClosestPoint)
// 				gizmoDelta = &delta
// 				targetGizmo.LastFrameClosestPoint = closestPointOnAxis
// 				targetGizmo.LastFrameMousePosition = mouseInput.Position
// 			}
// 		}
// 	}

// 	return gizmoDelta
// }

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

func setupRotationGizmo() *Gizmo {
	return &Gizmo{
		HoveredEntityID: -1,
		EntityIDToAxis: map[int]GizmoAxis{
			GizmoXDistancePickingID: GizmoAxis{DistanceBasedDelta: true},
			GizmoYDistancePickingID: GizmoAxis{DistanceBasedDelta: true},
			GizmoZDistancePickingID: GizmoAxis{DistanceBasedDelta: true},
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
