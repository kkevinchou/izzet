package izzet

import (
	"encoding/json"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/constants"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/collision/checks"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/kkevinchou/kitolib/utils"
)

var (
	maxCameraSpeed float64 = 400 // units per second
	slowSpeed      float64 = 100 // units per second
)

// Systems Context

func (g *Izzet) runCommandFrame(delta time.Duration) {
	frameInput := g.world.GetFrameInput()

	if frameInput.WindowEvent.Resized {
		w, h := g.window.GetSize()
		g.width, g.height = int(w), int(h)
		g.renderer.Resized(g.width, g.height)
	}

	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	if panels.DBG.EnableSpatialPartition {
		g.handleSpatialPartition()
	}

	if g.AppMode() == app.AppModePlay {
		for _, s := range g.playModeSystems {
			s.Update(delta, g.world)
		}
	} else if g.AppMode() == app.AppModeEditor {
		for _, s := range g.editorModeSystems {
			s.Update(delta, g.world)
		}
	}

	g.handleInputCommands(frameInput)

	if g.AppMode() == app.AppModeEditor {
		g.editorCameraMovement(frameInput, delta)
		g.handleGizmos(frameInput)
	}

	panels.DBG.CameraPosition = g.camera.Position
	panels.DBG.CameraOrientation = g.camera.Orientation
}

func (g *Izzet) handleSpatialPartition() {
	var spatialEntities []spatialpartition.Entity
	for _, entity := range g.world.Entities() {
		if !entity.HasBoundingBox() {
			continue
		}
		spatialEntities = append(spatialEntities, entity)
	}
	g.world.SpatialPartition().IndexEntities(spatialEntities)
}

var copiedEntity []byte

func (g *Izzet) handleInputCommands(frameInput input.Input) {
	mouseInput := frameInput.MouseInput
	// shutdown
	keyboardInput := frameInput.KeyboardInput
	if event, ok := keyboardInput[input.KeyboardKeyEscape]; ok && event.Event == input.KeyboardEventUp {
		if g.AppMode() == app.AppModeEditor {
			g.Shutdown()
		} else if g.AppMode() == app.AppModePlay {
			g.StopLiveWorld()
		}
	}

	if !InteractingWithUI() {
		if g.relativeMouseActive {
			g.platform.MoveMouse(g.relativeMouseOrigin[0], g.relativeMouseOrigin[1])
		}

		if mouseInput.MouseButtonEvent[1] == input.MouseButtonEventDown {
			g.relativeMouseActive = true
			g.relativeMouseOrigin[0] = int32(mouseInput.Position[0])
			g.relativeMouseOrigin[1] = int32(mouseInput.Position[1])
			g.platform.SetRelativeMouse(true)
		} else if mouseInput.MouseButtonEvent[1] == input.MouseButtonEventUp {
			g.relativeMouseActive = false
			g.platform.SetRelativeMouse(false)
		}
	}

	// undo/undo
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

	// delete entity
	if event, ok := keyboardInput[input.KeyboardKeyX]; ok {
		if event.Event == input.KeyboardEventUp {
			g.world.DeleteEntity(panels.SelectedEntity())
			panels.SelectEntity(nil)
		}
	}

	// copy entity
	if ctrlEvent, ok := keyboardInput[input.KeyboardKeyLCtrl]; ok {
		if ctrlEvent.Event == input.KeyboardEventDown {
			if cEvent, ok := keyboardInput[input.KeyboardKeyC]; ok {
				if cEvent.Event == input.KeyboardEventUp {
					if entity := panels.SelectedEntity(); entity != nil {
						var err error
						copiedEntity, err = json.Marshal(entity)
						if err != nil {
							panic(err)
						}
					}
				}
			}
		}
	}

	// paste entity
	if ctrlEvent, ok := keyboardInput[input.KeyboardKeyLCtrl]; ok {
		if ctrlEvent.Event == input.KeyboardEventDown {
			if vEvent, ok := keyboardInput[input.KeyboardKeyV]; ok {
				if vEvent.Event == input.KeyboardEventUp {
					var newEntity entities.Entity
					err := json.Unmarshal(copiedEntity, &newEntity)
					if err != nil {
						panic(err)
					}
					id := entities.GetNextIDAndAdvance()
					newEntity.ID = id

					g.world.AddEntity(&newEntity)
					panels.SelectEntity(&newEntity)
				}
			}
		}
	}

	// navmesh - move highlight
	if event, ok := keyboardInput[input.KeyboardKeyI]; ok {
		if event.Event == input.KeyboardEventUp {
			panels.DBG.VoxelHighlightZ--
			g.ResetNavMeshVAO()
		}
	}
	if event, ok := keyboardInput[input.KeyboardKeyK]; ok {
		if event.Event == input.KeyboardEventUp {
			panels.DBG.VoxelHighlightZ++
			g.ResetNavMeshVAO()
		}
	}
	if event, ok := keyboardInput[input.KeyboardKeyJ]; ok {
		if event.Event == input.KeyboardEventUp {
			panels.DBG.VoxelHighlightX--
			g.ResetNavMeshVAO()
		}
	}
	if event, ok := keyboardInput[input.KeyboardKeyL]; ok {
		if event.Event == input.KeyboardEventUp {
			panels.DBG.VoxelHighlightX++
			g.ResetNavMeshVAO()
		}
	}
}

func (g *Izzet) editorCameraMovement(frameInput input.Input, delta time.Duration) {
	mouseInput := frameInput.MouseInput
	keyboardInput := frameInput.KeyboardInput

	var viewRotation mgl64.Vec2
	var controlVector mgl64.Vec3
	if g.relativeMouseActive {
		var xRel, yRel float64
		var mouseSensitivity float64 = 0.003
		if mouseInput.Buttons[1] && !mouseInput.MouseMotionEvent.IsZero() {
			xRel += -mouseInput.MouseMotionEvent.XRel * mouseSensitivity
			yRel += -mouseInput.MouseMotionEvent.YRel * mouseSensitivity
		}
		viewRotation = mgl64.Vec2{xRel, yRel}
		controlVector = getControlVector(keyboardInput)
	}

	forwardVector := g.camera.Orientation.Rotate(mgl64.Vec3{0, 0, -1})
	upVector := g.camera.Orientation.Rotate(mgl64.Vec3{0, 1, 0})
	// there's probably away to get the right vector directly rather than going crossing the up vector :D
	rightVector := forwardVector.Cross(upVector)

	// calculate the quaternion for the delta in rotation
	deltaRotationX := mgl64.QuatRotate(viewRotation[1], rightVector)         // pitch
	deltaRotationY := mgl64.QuatRotate(viewRotation[0], mgl64.Vec3{0, 1, 0}) // yaw
	deltaRotation := deltaRotationY.Mul(deltaRotationX)

	newOrientation := deltaRotation.Mul(g.camera.Orientation)

	// don't let the camera go upside down
	if newOrientation.Rotate(mgl64.Vec3{0, 1, 0})[1] < 0 {
		newOrientation = g.camera.Orientation
	}

	g.camera.Orientation = newOrientation

	// keyboardInput := frameInput.KeyboardInput
	// controlVector := getControlVector(keyboardInput)
	if !frameInput.MouseInput.Buttons[1] {
		controlVector = mgl64.Vec3{}
	}

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

	if key, ok := keyboardInput[input.KeyboardKeyUp]; ok && key.Event == input.KeyboardEventDown {
		g.camera.Position = g.camera.Position.Add(forwardVector.Mul(slowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}
	if key, ok := keyboardInput[input.KeyboardKeyDown]; ok && key.Event == input.KeyboardEventDown {
		g.camera.Position = g.camera.Position.Add(forwardVector.Mul(-slowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}
	if key, ok := keyboardInput[input.KeyboardKeyLeft]; ok && key.Event == input.KeyboardEventDown {
		g.camera.Position = g.camera.Position.Add(rightVector.Mul(-slowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}
	if key, ok := keyboardInput[input.KeyboardKeyRight]; ok && key.Event == input.KeyboardEventDown {
		g.camera.Position = g.camera.Position.Add(rightVector.Mul(slowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}

	g.camera.LastFrameMovementVector = movementVector
}

func (g *Izzet) handleGizmos(frameInput input.Input) {
	mouseInput := frameInput.MouseInput

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

	var gizmoHovered bool = false
	// handle gizmo transforms
	if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
		entity := panels.SelectedEntity()
		delta := g.calculateGizmoDelta(gizmo.TGizmo, frameInput, entity.WorldPosition())
		if delta != nil {
			if entity.Parent != nil {
				// the computed position is in world space but entity.LocalPosition is in local space
				// to compute the new local space position we need to do conversions

				// compute the full transformation matrix, excluding local transformations
				// i.e. local transformations should not affect how the gizmo affects the entity
				transformMatrix := entities.ComputeParentAndJointTransformMatrix(entity)

				// take the new world position and convert it to local space
				worldPosition := entity.WorldPosition().Add(*delta)
				newPositionInLocalSpace := transformMatrix.Inv().Mul4x1(worldPosition.Vec4(1)).Vec3()

				entities.SetLocalPosition(entity, newPositionInLocalSpace)
			} else {
				entities.SetLocalPosition(entity, entity.LocalPosition.Add(*delta))
				// entities.SetLocalPosition(entity, *newPosition)
			}
		}

		gizmoHovered = gizmo.TGizmo.HoveredEntityID != -1
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
		entity := panels.SelectedEntity()
		newEntityRotation, hoverIndex := g.handleRotationGizmo(frameInput, panels.SelectedEntity())
		if newEntityRotation != nil {
			if entity.Parent != nil {
				transformMatrix := entities.ComputeParentAndJointTransformMatrix(entity)
				worldToLocalMatrix := transformMatrix.Inv()
				_, r, _ := utils.DecomposeF64(worldToLocalMatrix)
				computedRotation := r.Mul(*newEntityRotation)
				entities.SetLocalRotation(entity, computedRotation)
			} else {
				entities.SetLocalRotation(entity, *newEntityRotation)
			}
		}
		gizmoHovered = hoverIndex != -1
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
		scaleDelta, hovered := g.handleScaleGizmo(frameInput, panels.SelectedEntity())
		if scaleDelta != nil {
			entity := panels.SelectedEntity()
			scale := entities.GetLocalScale(entity)

			entities.SetScale(entity, scale.Add(*scaleDelta))
		}
		gizmoHovered = hovered
	}

	if !gizmoHovered && !InteractingWithUI() && mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
		entityID := g.renderer.GetEntityByPixelPosition(mouseInput.Position, g.height)
		if entityID == nil {
			panels.SelectEntity(nil)
			gizmo.CurrentGizmoMode = gizmo.GizmoModeNone
		} else {
			clickedEntity := g.world.GetEntityByID(*entityID)
			currentSelection := panels.SelectedEntity()

			if currentSelection != nil && currentSelection.ID != clickedEntity.ID {
				gizmo.CurrentGizmoMode = gizmo.GizmoModeNone
			}

			panels.SelectEntity(clickedEntity)
		}
	}

}

// TODO: move this method out of izzet and into the gizmo package?
func (g *Izzet) handleRotationGizmo(frameInput input.Input, selectedEntity *entities.Entity) (*mgl64.Quat, int) {
	if selectedEntity == nil {
		return nil, -1
	}

	mouseInput := frameInput.MouseInput
	nearPlanePos := g.mousePosToNearPlane(mouseInput, g.width, g.height)
	position := selectedEntity.WorldPosition()

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
			gizmo.R.ActivationRotation = selectedEntity.WorldRotation()
		}

		if !gizmo.R.Active {
			gizmo.R.HoverIndex = closestAxisIndex
		}
	} else if !gizmo.R.Active {
		// specifically check that the gizmo is not active before reseting.
		// this supports the scenario where we initially click and drag a gizmo
		// to the point where the mouse leaves the range of any axes
		gizmo.R.Reset()
		gizmo.R.HoverIndex = -1
	}

	if gizmo.R.Active && mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
		if gizmo.R.ActivationRotation != selectedEntity.WorldRotation() {
			g.AppendEdit(
				edithistory.NewRotationEdit(gizmo.R.ActivationRotation, selectedEntity.WorldRotation(), selectedEntity),
			)
		}
		gizmo.R.Reset()
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
		computedQuat := rotation.Mul(selectedEntity.WorldRotation())
		gizmo.R.MotionPivot = mouseInput.Position
		return &computedQuat, gizmo.R.HoverIndex
	}

	return nil, gizmo.R.HoverIndex
}

// TODO: move this method out of izzet and into the gizmo package?
func (g *Izzet) handleScaleGizmo(frameInput input.Input, selectedEntity *entities.Entity) (*mgl64.Vec3, bool) {
	if selectedEntity == nil {
		return nil, false
	}

	mouseInput := frameInput.MouseInput
	axisType := gizmo.NullAxis

	colorPickingID := g.renderer.GetEntityByPixelPosition(mouseInput.Position, g.height)
	if colorPickingID != nil {
		if *colorPickingID == constants.GizmoXAxisPickingID {
			axisType = gizmo.XAxis
		} else if *colorPickingID == constants.GizmoYAxisPickingID {
			axisType = gizmo.YAxis
		} else if *colorPickingID == constants.GizmoZAxisPickingID {
			axisType = gizmo.ZAxis
		} else if *colorPickingID == constants.GizmoAllAxisPickingID {
			axisType = gizmo.AllAxis
		} else {
			axisType = gizmo.NullAxis
			// we picked some other ID other than the translation gizmo
			colorPickingID = nil
		}
	}

	var scaleDir mgl64.Vec3
	if gizmo.S.HoveredAxisType == gizmo.XAxis {
		scaleDir = mgl64.Vec3{1, 0, 0}
	} else if gizmo.S.HoveredAxisType == gizmo.YAxis {
		scaleDir = mgl64.Vec3{0, 1, 0}
	} else if gizmo.S.HoveredAxisType == gizmo.ZAxis {
		scaleDir = mgl64.Vec3{0, 0, 1}
	}

	nearPlanePos := g.mousePosToNearPlane(mouseInput, g.width, g.height)
	position := selectedEntity.WorldPosition()
	// mouse is close to one of the axes, activate if we clicked
	if axisType != gizmo.NullAxis && mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
		if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(g.camera.Position, nearPlanePos, position, position.Add(scaleDir)); nonParallel {
			gizmo.S.OldClosestPoint = closestPointOnAxis
		}

		gizmo.S.Active = true
		gizmo.S.OldMousePosition = mouseInput.Position
		gizmo.S.HoveredAxisType = axisType
		gizmo.S.ActivationScale = entities.GetLocalScale(selectedEntity)
	}

	// reset if our gizmo isn't active
	if !gizmo.S.Active {
		gizmo.S.Reset()
		gizmo.S.HoveredAxisType = axisType
		return nil, gizmo.S.HoveredAxisType != gizmo.NullAxis
	}

	// if the gizmo was active and we receive a mouse up event, set it as inactive
	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
		scale := entities.GetLocalScale(selectedEntity)
		if gizmo.S.ActivationScale != scale {
			g.AppendEdit(
				edithistory.NewScaleEdit(gizmo.S.ActivationScale, scale, selectedEntity),
			)
		}
		gizmo.S.Reset()
	}

	var newEntityScale *mgl64.Vec3
	// handle the actual scaling of the entity
	if gizmo.S.HoveredAxisType == gizmo.AllAxis {
		// handle the all axes scaling

		delta := mouseInput.Position.Sub(gizmo.S.OldMousePosition)
		var sensitivity float64 = 0.005
		magnitude := (delta[0]*sensitivity - delta[1]*sensitivity)
		scale := mgl64.Vec3{1, 1, 1}.Mul(magnitude)
		newEntityScale = &scale
	} else if gizmo.S.HoveredAxisType != gizmo.NullAxis {

		// handle when mouse moves the translation slider
		if mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
			if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(g.camera.Position, nearPlanePos, position, position.Add(scaleDir)); nonParallel {
				newScale := (closestPointOnAxis.Sub(gizmo.S.OldClosestPoint)).Mul(0.05)
				gizmo.S.OldClosestPoint = closestPointOnAxis
				newEntityScale = &newScale
			}
		}
	}

	gizmo.S.OldMousePosition = mouseInput.Position
	return newEntityScale, gizmo.S.HoveredAxisType != gizmo.NullAxis
}

// TODO: move this method out of izzet and into the gizmo package?
func (g *Izzet) calculateGizmoDelta(targetGizmo *gizmo.Gizmo, frameInput input.Input, position mgl64.Vec3) *mgl64.Vec3 {
	mouseInput := frameInput.MouseInput

	colorPickingID := g.renderer.GetEntityByPixelPosition(mouseInput.Position, g.height)

	if colorPickingID != nil {
		if _, ok := targetGizmo.EntityIDToAxis[*colorPickingID]; ok {
			if targetGizmo.HoveredEntityID == -1 {
				targetGizmo.HoveredEntityID = *colorPickingID
			}
		} else {
			colorPickingID = nil
		}
	}

	nearPlanePos := g.mousePosToNearPlane(mouseInput, g.width, g.height)
	if colorPickingID != nil {
		if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
			axis := targetGizmo.EntityIDToAxis[*colorPickingID]
			if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(g.camera.Position, nearPlanePos, position, position.Add(axis.Direction)); nonParallel {
				targetGizmo.LastFrameClosestPoint = closestPointOnAxis
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
		return nil
	}

	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
		targetGizmo.Reset()
	}

	var gizmoDelta *mgl64.Vec3

	if mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
		axis := targetGizmo.EntityIDToAxis[targetGizmo.HoveredEntityID]
		if _, closestPointOnAxis, nonParallel := checks.ClosestPointsInfiniteLines(g.camera.Position, nearPlanePos, position, position.Add(axis.Direction)); nonParallel {
			delta := closestPointOnAxis.Sub(targetGizmo.LastFrameClosestPoint)
			gizmoDelta = &delta
			targetGizmo.LastFrameClosestPoint = closestPointOnAxis
		}
	}

	return gizmoDelta
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

func InteractingWithUI() bool {
	anyPopup := imgui.IsPopupOpenV("", imgui.PopupFlagsAnyPopup)
	anyWindow := imgui.IsWindowHoveredV(imgui.HoveredFlagsAnyWindow)
	return anyPopup || anyWindow
}
