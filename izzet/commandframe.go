package izzet

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/collision/checks"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/input"
)

var (
	maxCameraSpeed float64 = 400 // units per second
)

func (g *Izzet) runCommandFrame(frameInput input.Input, delta time.Duration) {
	for _, entity := range g.Entities() {
		if entity.AnimationPlayer != nil {
			entity.AnimationPlayer.Update(delta)
		}
	}

	keyboardInput := frameInput.KeyboardInput
	if _, ok := keyboardInput[input.KeyboardKeyEscape]; ok {
		g.Shutdown()
	}

	if _, ok := keyboardInput[input.KeyboardKeyLCtrl]; ok {
		if _, ok := keyboardInput[input.KeyboardKeyLShift]; ok {
			if event, ok := keyboardInput[input.KeyboardKeyZ]; ok {
				if event.Event == input.KeyboardEventUp {
					g.Redo()
				}
			}
		} else {
			if event, ok := keyboardInput[input.KeyboardKeyZ]; ok {
				if event.Event == input.KeyboardEventUp {
					g.Undo()
				}
			}
		}
	}

	mouseInput := frameInput.MouseInput
	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
		if newSelection := g.selectEntity(frameInput); newSelection {
			// fmt.Println("RESET")
			gizmo.T.Reset()
			gizmo.R.Reset()
			gizmo.S.Reset()
			gizmo.CurrentGizmoMode = gizmo.GizmoModeNone
		}
	}
	g.cameraMovement(frameInput, delta)

	// set gizmo mode
	if panels.SelectedEntity() != nil {
		if gizmo.CurrentGizmoMode == gizmo.GizmoModeNone {
			gizmo.CurrentGizmoMode = gizmo.GizmoModeTranslation
		}
		keyboardInput := frameInput.KeyboardInput
		if _, ok := keyboardInput[input.KeyboardKeyT]; ok {
			gizmo.CurrentGizmoMode = gizmo.GizmoModeTranslation
		} else if _, ok := keyboardInput[input.KeyboardKeyR]; ok {
			gizmo.CurrentGizmoMode = gizmo.GizmoModeRotation
		} else if _, ok := keyboardInput[input.KeyboardKeyE]; ok {
			gizmo.CurrentGizmoMode = gizmo.GizmoModeScale
		}
	}

	// handle gizmo transforms
	if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
		newEntityPosition := g.handleTranslationGizmo(frameInput, panels.SelectedEntity())
		if newEntityPosition != nil {
			panels.SelectedEntity().Position = *newEntityPosition
		}
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
		newEntityRotation := g.handleRotationGizmo(frameInput, panels.SelectedEntity())
		if newEntityRotation != nil {
			panels.SelectedEntity().Rotation = *newEntityRotation
		}
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
		scaleDelta := g.handleScaleGizmo(frameInput, panels.SelectedEntity())
		if scaleDelta != nil {
			entity := panels.SelectedEntity()
			entity.Scale = entity.Scale.Add(*scaleDelta)
		}
	}
}

func (g *Izzet) selectEntity(frameInput input.Input) bool {
	mouseInput := frameInput.MouseInput

	var newSelection bool
	// select the entity in the hierarchy
	entityID := g.renderer.GetEntityByPixelPosition(mouseInput.Position)
	if entityID == nil {
		newSelection = panels.SelectEntity(nil)
		gizmo.CurrentGizmoMode = gizmo.GizmoModeNone
	} else {
		for _, e := range g.Entities() {
			if e.ID == *entityID {
				newSelection = panels.SelectEntity(e)
			}
		}
	}

	return newSelection
}

func (g *Izzet) cameraMovement(frameInput input.Input, delta time.Duration) {
	if !frameInput.MouseInput.Buttons[1] {
		return
	}

	var xRel, yRel float64
	mouseInput := frameInput.MouseInput
	var mouseSensitivity float64 = 0.003
	if mouseInput.Buttons[1] && !mouseInput.MouseMotionEvent.IsZero() {
		xRel += -mouseInput.MouseMotionEvent.XRel * mouseSensitivity
		yRel += -mouseInput.MouseMotionEvent.YRel * mouseSensitivity
	}
	forwardVector := g.camera.Orientation.Rotate(mgl64.Vec3{0, 0, -1})
	upVector := g.camera.Orientation.Rotate(mgl64.Vec3{0, 1, 0})
	// there's probably away to get the right vector directly rather than going crossing the up vector :D
	rightVector := forwardVector.Cross(upVector)

	// calculate the quaternion for the delta in rotation
	deltaRotationX := mgl64.QuatRotate(yRel, rightVector)         // pitch
	deltaRotationY := mgl64.QuatRotate(xRel, mgl64.Vec3{0, 1, 0}) // yaw
	deltaRotation := deltaRotationY.Mul(deltaRotationX)

	newOrientation := deltaRotation.Mul(g.camera.Orientation)

	// don't let the camera go upside down
	if newOrientation.Rotate(mgl64.Vec3{0, 1, 0})[1] < 0 {
		newOrientation = g.camera.Orientation
	}

	g.camera.Orientation = newOrientation

	keyboardInput := frameInput.KeyboardInput
	controlVector := getControlVector(keyboardInput)

	movementVector := rightVector.Mul(controlVector[0]).Add(mgl64.Vec3{0, 1, 0}.Mul(controlVector[1])).Add(forwardVector.Mul(controlVector[2]))

	if !movementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
		if g.camera.LastFrameMovementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
			// this is the starting speed that the camera accelerates from
			g.camera.Speed = maxCameraSpeed * 0.3
		} else {
			// TODO(kevin) parameterize how slowly we accelerate based on how long we want to drift for
			g.camera.Speed *= 1.03
			if g.camera.Speed > maxCameraSpeed {
				g.camera.Speed = maxCameraSpeed
			}
		}
	}

	if !movementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
		movementVector = movementVector.Normalize()
	}

	perFrameMovement := float64(g.camera.Speed) * float64(delta.Milliseconds()) / 1000
	movementDelta := movementVector.Mul(perFrameMovement)

	if movementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
		// start drifting if we were moving last frame but not the current one
		if !g.camera.LastFrameMovementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
			g.camera.Drift = g.camera.LastFrameMovementVector.Mul(perFrameMovement)
		} else {
			// TODO(kevin) parameterize how slowly we decay based on how long we want to drift for
			g.camera.Drift = g.camera.Drift.Mul(0.93)
			if g.camera.Drift.Len() < 0.01 {
				g.camera.Drift = mgl64.Vec3{}
			}
		}
		g.camera.Speed = 0
	} else {
		// if we're actively moving the camera, remove all drift
		g.camera.Drift = mgl64.Vec3{}
	}

	g.camera.Position = g.camera.Position.Add(movementDelta).Add(g.camera.Drift)
	g.camera.LastFrameMovementVector = movementVector
}

// TODO: move this method out of izzet and into the gizmo package?
func (g *Izzet) handleRotationGizmo(frameInput input.Input, selectedEntity *entities.Entity) *mgl64.Quat {
	if selectedEntity == nil {
		return nil
	}

	mouseInput := frameInput.MouseInput
	nearPlanePos := g.mousePosToNearPlane(mouseInput)
	position := selectedEntity.Position

	var minDist *float64
	closestAxisIndex := -1

	for i, axis := range gizmo.R.Axes {
		ray := collider.Ray{Origin: g.camera.Position, Direction: nearPlanePos.Sub(g.camera.Position).Normalize()}
		plane := collider.Plane{Point: position, Normal: axis.Normal}

		intersect, front := checks.IntersectRayPlane(ray, plane)
		if !front || intersect == nil {
			continue
		}

		dist := position.Sub(*intersect).Len()

		circleRadius := axis.Radius
		activationRange := 3

		if dist >= float64(circleRadius)-float64(activationRange) && dist <= float64(circleRadius)+float64(activationRange) {
		} else {
			continue
		}

		if minDist == nil || dist < *minDist {
			minDist = &dist
			closestAxisIndex = i
		}
	}

	// mouse is close to one of the axes
	if minDist != nil {
		if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
			gizmo.R.Active = true
			gizmo.R.MotionPivot = mouseInput.Position
			gizmo.R.HoverIndex = closestAxisIndex
			gizmo.R.ActivationRotation = selectedEntity.Rotation
			// fmt.Println("Activate ID Rotation", selectedEntity.ID)
		}

		if !gizmo.R.Active {
			gizmo.R.HoverIndex = closestAxisIndex
		}
	} else if !gizmo.R.Active {
		gizmo.R.HoverIndex = -1
	}

	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
		gizmo.R.Active = false
		gizmo.R.HoverIndex = closestAxisIndex
		if gizmo.R.ActivationRotation != selectedEntity.Rotation {
			// fmt.Println("Edit ID Rotation", selectedEntity.ID)
			g.AppendEdit(
				edithistory.NewRotationEdit(gizmo.R.ActivationRotation, selectedEntity.Rotation, selectedEntity),
			)
		}
	}

	// handle when mouse moves the rotation gizmo
	if gizmo.R.Active && mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
		viewDir := g.Camera().Orientation.Rotate(mgl64.Vec3{0, 0, -1})
		delta := mouseInput.Position.Sub(gizmo.R.MotionPivot)
		sensitivity := 2 * math.Pi / 1000
		rotation := mgl64.QuatIdent()

		if gizmo.R.HoverIndex == 0 {
			// rotation around Z axis
			horizontalAlignment := math.Abs(g.Camera().Position.Sub(position).Normalize().Dot(mgl64.Vec3{0, 0, -1}))
			magnitude := (horizontalAlignment*delta[0] + delta[1]) * float64(sensitivity)
			var dir float64 = 1
			if viewDir.Dot(mgl64.Vec3{0, 0, -1}) > 0 {
				dir = -1
			}
			rotation = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, 0, dir})
		} else if gizmo.R.HoverIndex == 1 {
			// rotation around X axis
			horizontalAlignment := math.Abs(g.Camera().Position.Sub(position).Normalize().Dot(mgl64.Vec3{1, 0, 0}))
			magnitude := (horizontalAlignment*delta[0] + delta[1]) * float64(sensitivity)
			var dir float64 = 1
			if viewDir.Dot(mgl64.Vec3{-1, 0, 0}) > 0 {
				dir = -1
			}
			rotation = mgl64.QuatRotate(magnitude, mgl64.Vec3{dir, 0, 0})
		} else if gizmo.R.HoverIndex == 2 {
			// rotation around Y axis
			verticalAlignment := math.Abs(g.Camera().Position.Sub(position).Normalize().Dot(mgl64.Vec3{0, 1, 0}))
			magnitude := (delta[0] + verticalAlignment*delta[1]) * float64(sensitivity)
			var dir float64 = 1
			if viewDir.Dot(mgl64.Vec3{0, -1, 0}) > 0 {
				dir = -1
			}
			rotation = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, dir, 0})
		}
		computedQuat := rotation.Mul(selectedEntity.Rotation)
		gizmo.R.MotionPivot = mouseInput.Position
		return &computedQuat
	}

	return nil
}

// TODO: move this method out of izzet and into the gizmo package?
func (g *Izzet) handleScaleGizmo(frameInput input.Input, selectedEntity *entities.Entity) *mgl64.Vec3 {
	if selectedEntity == nil {
		return nil
	}

	mouseInput := frameInput.MouseInput
	nearPlanePos := g.mousePosToNearPlane(mouseInput)
	position := selectedEntity.Position

	var minDist *float64
	minAxis := mgl64.Vec3{}
	closestAxisIndex := -1

	for i, axis := range gizmo.S.Axes {
		if a, b, nonParallel := checks.ClosestPointsInfiniteLineVSLine(g.camera.Position, nearPlanePos, position, position.Add(axis)); nonParallel {
			length := a.Sub(b).Len()
			if length > gizmo.ActivationRadius {
				continue
			}

			if minDist == nil || length < *minDist {
				minAxis = axis
				minDist = &length
				closestAxisIndex = i
			}
		}
	}

	// mouse is close to one of the axes
	if minDist != nil {
		if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
			gizmo.S.Active = true
			gizmo.S.ScaleDir = minAxis
			gizmo.S.MotionPivot = mouseInput.Position
			gizmo.S.HoverIndex = closestAxisIndex
			gizmo.S.ActivationScale = selectedEntity.Scale
			// fmt.Println("Activate ID Scale", selectedEntity.ID)
		}

		if !gizmo.S.Active {
			gizmo.S.HoverIndex = closestAxisIndex
		}
	} else if !gizmo.S.Active {
		gizmo.S.HoverIndex = -1
	}

	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
		gizmo.S.Active = false
		gizmo.S.HoverIndex = closestAxisIndex
		// fmt.Println("Edit ID Scale", selectedEntity.ID)
		if gizmo.S.ActivationScale != selectedEntity.Scale {
			g.AppendEdit(
				edithistory.NewScaleEdit(gizmo.S.ActivationScale, selectedEntity.Scale, selectedEntity),
			)
		}
	}

	var newEntityScale *mgl64.Vec3
	// handle when mouse moves the translation slider
	if gizmo.S.Active && mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
		if _, _, nonParallel := checks.ClosestPointsInfiniteLines(g.camera.Position, nearPlanePos, position, position.Add(gizmo.S.ScaleDir)); nonParallel {
			viewDir := g.Camera().Orientation.Rotate(mgl64.Vec3{0, 0, -1})
			delta := mouseInput.Position.Sub(gizmo.S.MotionPivot)
			sensitivity := 0.01
			magnitude := (delta[0] - delta[1]) * float64(sensitivity)

			if gizmo.S.HoverIndex == 0 {
				// X Scale
				var dir float64 = 1
				if viewDir.Dot(mgl64.Vec3{0, 0, -1}) < 0 {
					dir = -1
				}
				newEntityScale = &mgl64.Vec3{dir * magnitude, 0, 0}
			} else if gizmo.S.HoverIndex == 1 {
				// Y Scale
				var dir float64 = 1
				// if viewDir.Dot(mgl64.Vec3{0, 1, 0}) > 0 {
				// 	dir = -1
				// }
				newEntityScale = &mgl64.Vec3{0, dir * magnitude, 0}
			} else if gizmo.S.HoverIndex == 2 {
				// Z Scale
				var dir float64 = 1
				if viewDir.Dot(mgl64.Vec3{0, 0, 1}) < 0 {
					dir = -1
				}
				newEntityScale = &mgl64.Vec3{0, 0, dir * magnitude}
			}
		}
		gizmo.S.MotionPivot = mouseInput.Position
	}

	return newEntityScale
}

// TODO: move this method out of izzet and into the gizmo package?
func (g *Izzet) handleTranslationGizmo(frameInput input.Input, selectedEntity *entities.Entity) *mgl64.Vec3 {
	if selectedEntity == nil {
		return nil
	}

	mouseInput := frameInput.MouseInput
	nearPlanePos := g.mousePosToNearPlane(mouseInput)
	position := selectedEntity.Position

	var minDist *float64
	minAxis := mgl64.Vec3{}
	motionPivot := mgl64.Vec3{}
	closestAxisIndex := -1

	for i, axis := range gizmo.T.Axes {
		if a, b, nonParallel := checks.ClosestPointsInfiniteLineVSLine(g.camera.Position, nearPlanePos, position, position.Add(axis)); nonParallel {
			length := a.Sub(b).Len()
			if length > gizmo.ActivationRadius {
				continue
			}

			if minDist == nil || length < *minDist {
				minAxis = axis
				minDist = &length
				motionPivot = b
				closestAxisIndex = i
			}
		}
	}

	// mouse is close to one of the axes
	if minDist != nil {
		if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
			gizmo.T.Active = true
			gizmo.T.TranslationDir = minAxis
			gizmo.T.MotionPivot = motionPivot.Sub(position)
			gizmo.T.HoverIndex = closestAxisIndex
			gizmo.T.ActivationPosition = position
			// fmt.Println("Activate ID translate", selectedEntity.ID)
		}

		if !gizmo.T.Active {
			gizmo.T.HoverIndex = closestAxisIndex
		}
	} else if !gizmo.T.Active {
		gizmo.T.HoverIndex = -1
	}

	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
		gizmo.T.Active = false
		gizmo.T.HoverIndex = closestAxisIndex
		// fmt.Println("Edit ID translate", selectedEntity.ID)
		if gizmo.T.ActivationPosition != position {
			g.AppendEdit(
				edithistory.NewPositionEdit(gizmo.T.ActivationPosition, selectedEntity.Position, selectedEntity),
			)
		}
	}

	var newEntityPosition *mgl64.Vec3
	// handle when mouse moves the translation slider
	if gizmo.T.Active && mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
		if _, b, nonParallel := checks.ClosestPointsInfiniteLines(g.camera.Position, nearPlanePos, position, position.Add(gizmo.T.TranslationDir)); nonParallel {
			newPosition := b.Sub(gizmo.T.MotionPivot)
			newPosition[0] = float64(int(newPosition[0]))
			newPosition[1] = float64(int(newPosition[1]))
			newPosition[2] = float64(int(newPosition[2]))
			newEntityPosition = &newPosition
		}
	}

	return newEntityPosition
}

func getControlVector(keyboardInput input.KeyboardInput) mgl64.Vec3 {
	var controlVector mgl64.Vec3
	if key, ok := keyboardInput[input.KeyboardKeyW]; ok && key.Event == input.KeyboardEventDown {
		controlVector[2]++
	}
	if key, ok := keyboardInput[input.KeyboardKeyS]; ok && key.Event == input.KeyboardEventDown {
		controlVector[2]--
	}
	if key, ok := keyboardInput[input.KeyboardKeyA]; ok && key.Event == input.KeyboardEventDown {
		controlVector[0]--
	}
	if key, ok := keyboardInput[input.KeyboardKeyD]; ok && key.Event == input.KeyboardEventDown {
		controlVector[0]++
	}
	if key, ok := keyboardInput[input.KeyboardKeyLShift]; ok && key.Event == input.KeyboardEventDown {
		controlVector[1]--
	}
	if key, ok := keyboardInput[input.KeyboardKeySpace]; ok && key.Event == input.KeyboardEventDown {
		controlVector[1]++
	}
	return controlVector
}
