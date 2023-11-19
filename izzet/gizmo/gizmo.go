package gizmo

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/collision/checks"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/input"
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
	GizmoXZAxisPickingID  int = 1000000004
	GizmoXYAxisPickingID  int = 1000000005
	GizmoYZAxisPickingID  int = 1000000006

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

	AccumulatedDelta mgl64.Vec3
	LastSnapVector   mgl64.Vec3

	ActivationPosition mgl64.Vec3
	ActivationScale    mgl64.Vec3
	ActivationRotation mgl64.Quat
}

type GizmoEvent string

const (
	GizmoEventActivated GizmoEvent = "ACTIVATED"
	GizmoEventCompleted GizmoEvent = "COMPLETED"
	GizmoEventNone      GizmoEvent = "NONE"
)

type Positionable interface {
	Position() mgl64.Vec3
}

func CalculateGizmoDelta(targetGizmo *Gizmo, frameInput input.Input, mousePosition mgl64.Vec2, gizmoPosition mgl64.Vec3, cameraPosition mgl64.Vec3, nearPlanePosition mgl64.Vec3, hoveredEntityID *int, snapSize int) (*mgl64.Vec3, GizmoEvent) {
	gizmoEvent := GizmoEventNone
	startStatus := targetGizmo.Active
	mouseInput := frameInput.MouseInput

	if hoveredEntityID != nil {
		if _, ok := targetGizmo.EntityIDToAxis[*hoveredEntityID]; ok {
			if !mouseInput.Buttons[0] {
				targetGizmo.HoveredEntityID = *hoveredEntityID
			}
		} else {
			hoveredEntityID = nil
		}
	}

	if hoveredEntityID != nil {
		if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
			axis := targetGizmo.EntityIDToAxis[*hoveredEntityID]
			if axis.DistanceBasedDelta {
				targetGizmo.LastFrameMousePosition = mousePosition
			} else if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(cameraPosition, nearPlanePosition, gizmoPosition, gizmoPosition.Add(axis.Direction)); nonParallel {
				targetGizmo.LastFrameClosestPoint = closestPointOnAxis
				targetGizmo.LastFrameMousePosition = mousePosition
			} else if !nonParallel && (*hoveredEntityID == GizmoAllAxisPickingID) {
				targetGizmo.LastFrameClosestPoint = closestPointOnAxis
				targetGizmo.LastFrameMousePosition = mousePosition
			} else if !nonParallel && (*hoveredEntityID == GizmoXZAxisPickingID || *hoveredEntityID == GizmoXYAxisPickingID || *hoveredEntityID == GizmoYZAxisPickingID) {
				plane := planeFromAxis(targetGizmo.HoveredEntityID, gizmoPosition)
				ray := collider.Ray{Origin: cameraPosition, Direction: nearPlanePosition.Sub(cameraPosition).Normalize()}

				position, hit := checks.IntersectRayPlane(ray, plane)
				if hit {
					targetGizmo.LastFrameClosestPoint = position
					targetGizmo.LastFrameMousePosition = mousePosition
				}
			} else {
				panic("parallel")
			}

			targetGizmo.Active = true
		}
	} else if !targetGizmo.Active {
		// specifically check that the gizmo is not active before reseting.
		// this supports the scenario where we initially click and drag a gizmo
		// to the point where the mouse leaves the range of any axes
		targetGizmo.Reset()
	}

	if !targetGizmo.Active {
		return nil, gizmoEvent
	}

	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
		targetGizmo.Reset()
	}

	var gizmoDelta *mgl64.Vec3

	if mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
		axis := targetGizmo.EntityIDToAxis[targetGizmo.HoveredEntityID]

		if axis.DistanceBasedDelta {
			// mouse position based deltas, store the x,y mouse delta in the return value with 0 for the z value
			delta := mousePosition.Sub(targetGizmo.LastFrameMousePosition).Vec3(0)
			targetGizmo.LastFrameMousePosition = mousePosition

			targetGizmo.AccumulatedDelta = targetGizmo.AccumulatedDelta.Add(delta)
			calculatedPosition := targetGizmo.LastSnapVector.Add(targetGizmo.AccumulatedDelta)

			snappedXPosition := math.Trunc(calculatedPosition.X()/float64(snapSize)) * float64(snapSize)
			snappedYPosition := math.Trunc(calculatedPosition.Y()/float64(snapSize)) * float64(snapSize)
			snappedZPosition := math.Trunc(calculatedPosition.Z()/float64(snapSize)) * float64(snapSize)

			var snappedDelta mgl64.Vec3
			if math.Trunc(targetGizmo.LastSnapVector.X()) != snappedXPosition {
				snapDeltaX := float64(snappedXPosition) - targetGizmo.LastSnapVector.X()
				snappedDelta[0] = snapDeltaX

				gizmoDelta = &snappedDelta
				targetGizmo.AccumulatedDelta[0] -= snapDeltaX
				targetGizmo.LastSnapVector[0] += snapDeltaX
			}
			if math.Trunc(targetGizmo.LastSnapVector.Y()) != snappedYPosition {
				snapDeltaY := float64(snappedYPosition) - targetGizmo.LastSnapVector.Y()
				snappedDelta[1] = snapDeltaY

				gizmoDelta = &snappedDelta
				targetGizmo.AccumulatedDelta[1] -= snapDeltaY
				targetGizmo.LastSnapVector[1] += snapDeltaY
			}
			if math.Trunc(targetGizmo.LastSnapVector.Z()) != snappedZPosition {
				snapDeltaZ := float64(snappedZPosition) - targetGizmo.LastSnapVector.Z()
				snappedDelta[2] = snapDeltaZ

				gizmoDelta = &snappedDelta
				targetGizmo.AccumulatedDelta[2] -= snapDeltaZ
				targetGizmo.LastSnapVector[2] += snapDeltaZ
			}
		} else if targetGizmo.HoveredEntityID == GizmoAllAxisPickingID {
			mouseDelta := mousePosition.Sub(targetGizmo.LastFrameMousePosition)
			magnitude := (mouseDelta[0] - mouseDelta[1])
			delta := mgl64.Vec3{1, 1, 1}.Mul(magnitude)
			gizmoDelta = &delta
			targetGizmo.LastFrameMousePosition = mousePosition
		} else if targetGizmo.HoveredEntityID == GizmoXZAxisPickingID || targetGizmo.HoveredEntityID == GizmoXYAxisPickingID || targetGizmo.HoveredEntityID == GizmoYZAxisPickingID {
			plane := planeFromAxis(targetGizmo.HoveredEntityID, gizmoPosition)
			ray := collider.Ray{Origin: cameraPosition, Direction: nearPlanePosition.Sub(cameraPosition).Normalize()}

			position, hit := checks.IntersectRayPlane(ray, plane)
			if hit {
				delta := position.Sub(targetGizmo.LastFrameClosestPoint)
				targetGizmo.LastFrameClosestPoint = position

				gizmoDelta = &delta
			}
		} else {
			if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(cameraPosition, nearPlanePosition, gizmoPosition, gizmoPosition.Add(axis.Direction)); nonParallel {
				delta := closestPointOnAxis.Sub(targetGizmo.LastFrameClosestPoint)
				targetGizmo.LastFrameClosestPoint = closestPointOnAxis

				targetGizmo.AccumulatedDelta = targetGizmo.AccumulatedDelta.Add(delta)
				calculatedPosition := targetGizmo.LastSnapVector.Add(targetGizmo.AccumulatedDelta)

				snappedXPosition := math.Trunc(calculatedPosition.X()/float64(snapSize)) * float64(snapSize)
				snappedYPosition := math.Trunc(calculatedPosition.Y()/float64(snapSize)) * float64(snapSize)
				snappedZPosition := math.Trunc(calculatedPosition.Z()/float64(snapSize)) * float64(snapSize)

				var snappedDelta mgl64.Vec3
				if math.Trunc(targetGizmo.LastSnapVector.X()) != snappedXPosition {
					snapDeltaX := float64(snappedXPosition) - targetGizmo.LastSnapVector.X()
					snappedDelta[0] = snapDeltaX

					gizmoDelta = &snappedDelta
					targetGizmo.AccumulatedDelta[0] -= snapDeltaX
					targetGizmo.LastSnapVector[0] += snapDeltaX
				}
				if math.Trunc(targetGizmo.LastSnapVector.Y()) != snappedYPosition {
					snapDeltaY := float64(snappedYPosition) - targetGizmo.LastSnapVector.Y()
					snappedDelta[1] = snapDeltaY

					gizmoDelta = &snappedDelta
					targetGizmo.AccumulatedDelta[1] -= snapDeltaY
					targetGizmo.LastSnapVector[1] += snapDeltaY
				}
				if math.Trunc(targetGizmo.LastSnapVector.Z()) != snappedZPosition {
					snapDeltaZ := float64(snappedZPosition) - targetGizmo.LastSnapVector.Z()
					snappedDelta[2] = snapDeltaZ

					gizmoDelta = &snappedDelta
					targetGizmo.AccumulatedDelta[2] -= snapDeltaZ
					targetGizmo.LastSnapVector[2] += snapDeltaZ
				}
			}
		}
	}

	endStatus := targetGizmo.Active

	if startStatus == false && endStatus == true {
		gizmoEvent = GizmoEventActivated
	} else if startStatus == true && endStatus == false {
		gizmoEvent = GizmoEventCompleted
		targetGizmo.AccumulatedDelta = mgl64.Vec3{}
	}

	return gizmoDelta, gizmoEvent
}

func setupTranslationGizmo() *Gizmo {
	return &Gizmo{
		HoveredEntityID: -1,
		EntityIDToAxis: map[int]GizmoAxis{
			GizmoXAxisPickingID:  GizmoAxis{Direction: mgl64.Vec3{1, 0, 0}},
			GizmoYAxisPickingID:  GizmoAxis{Direction: mgl64.Vec3{0, 1, 0}},
			GizmoZAxisPickingID:  GizmoAxis{Direction: mgl64.Vec3{0, 0, 1}},
			GizmoXZAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{}},
			GizmoXYAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{}},
			GizmoYZAxisPickingID: GizmoAxis{Direction: mgl64.Vec3{}},
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

func planeFromAxis(hoveredEntityID int, gizmoPosition mgl64.Vec3) collider.Plane {
	var plane collider.Plane
	if hoveredEntityID == GizmoXZAxisPickingID {
		plane = collider.Plane{Point: gizmoPosition, Normal: mgl64.Vec3{0, 1, 0}}
	} else if hoveredEntityID == GizmoXYAxisPickingID {
		plane = collider.Plane{Point: gizmoPosition, Normal: mgl64.Vec3{0, 0, 1}}
	} else if hoveredEntityID == GizmoYZAxisPickingID {
		plane = collider.Plane{Point: gizmoPosition, Normal: mgl64.Vec3{1, 0, 0}}
	}
	return plane
}
