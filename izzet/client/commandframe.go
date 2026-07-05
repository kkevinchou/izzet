package client

import (
	"fmt"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/appmode"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/client/edithistory"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
)

// Systems Context

func (g *Client) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	frameInput := g.GetFrameInput()

	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	g.world.ReindexSpatialEntities()

	if g.AppMode() == appmode.Play {
		for _, s := range g.playModeSystems {
			s.Update(delta, g.world)
		}
	} else if g.AppMode() == appmode.Editor {
		for _, s := range g.editorModeSystems {
			s.Update(delta, g.world)
		}
	}

	g.handleCommonInputCommands(frameInput)

	if g.AppMode() == appmode.Editor {
		g.handleEditorInputCommands(frameInput)
		g.editorCameraMovement(frameInput, delta)
		g.handleGizmos(frameInput)
	}

	if g.MouseCaptured() {
		g.platform.MoveMouse(g.capturedMouseOrigin[0], g.capturedMouseOrigin[1])
	}
	g.RuntimeConfig().CameraPosition = g.camera.Position
	g.RuntimeConfig().CameraRotation = g.camera.Rotation
}

var copiedEntity []byte

func (g *Client) handleEditorInputCommands(frameInput input.Input) {
	// mouseInput := frameInput.MouseInput
	keyboardInput := frameInput.KeyboardInput

	if _, ok := keyboardInput[input.KeyboardKeyF5]; ok {
		err := g.Connect()
		if err != nil {
			fmt.Println(err)
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
			selectedEntity := g.SelectedEntity()
			if selectedEntity != nil {
				g.world.DeleteEntity(selectedEntity.ID)
			}
			g.SelectEntity(nil)
		}
	}

	// copy entity
	if _, ok := keyboardInput[input.KeyboardKeyLCtrl]; ok {
		if cEvent, ok := keyboardInput[input.KeyboardKeyC]; ok {
			if cEvent.Event == input.KeyboardEventUp {
				if e := g.SelectedEntity(); e != nil {
					var err error
					copiedEntity, err = serialization.SerializeEntity(e)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}

	// paste entity
	if _, ok := keyboardInput[input.KeyboardKeyLCtrl]; ok {
		if vEvent, ok := keyboardInput[input.KeyboardKeyV]; ok {
			if vEvent.Event == input.KeyboardEventUp {
				e, err := serialization.DeserializeEntity(copiedEntity, g.AssetManager())
				if err == nil {
					id := entity.GetNextIDAndAdvance()
					e.ID = id
					g.world.AddEntity(e)
					g.SelectEntity(e)
				}
			}
		}
	}

	mouseInput := frameInput.MouseInput
	if g.renderSystem.GameWindowHovered() {
		if mouseInput.MouseButtonEvent[1] == input.MouseButtonEventDown {
			g.capturedMouseOrigin[0] = int32(mouseInput.Position[0])
			g.capturedMouseOrigin[1] = int32(mouseInput.Position[1])
			g.SetMouseCaptured(true)
		} else if mouseInput.MouseButtonEvent[1] == input.MouseButtonEventUp {
			g.SetMouseCaptured(false)
		}
	}
}

func (g *Client) handleCommonInputCommands(frameInput input.Input) {
	for _, cmd := range frameInput.Commands {
		if c, ok := cmd.(input.FileDropCommand); ok {
			fmt.Println("received drop file command", c.File)
		} else if _, ok := cmd.(input.QuitCommand); ok {
			g.Shutdown()
		}
	}

	keyboardInput := frameInput.KeyboardInput

	if event, ok := keyboardInput[input.KeyboardKeyF11]; ok && event.Event == input.KeyboardEventUp {
		g.ConfigureUI(!g.runtimeConfig.UIEnabled)
	}

	// shutdown
	if event, ok := keyboardInput[input.KeyboardKeyEscape]; ok && event.Event == input.KeyboardEventUp {
		if g.AppMode() == appmode.Editor {
			g.Shutdown()
		} else if g.AppMode() == appmode.Play {
			g.DisconnectClient()
		}
	}
}

func (g *Client) editorCameraMovement(frameInput input.Input, delta time.Duration) {
	mouseInput := frameInput.MouseInput
	keyboardInput := frameInput.KeyboardInput

	var viewRotation mgl64.Vec2
	var controlVector mgl64.Vec3

	if g.MouseCaptured() {
		var xRel, yRel float64
		var mouseSensitivity float64 = 0.003
		if mouseInput.MouseButtonState[1] && !mouseInput.MouseMotionEvent.IsZero() {
			xRel += -mouseInput.MouseMotionEvent.XRel * mouseSensitivity
			yRel += -mouseInput.MouseMotionEvent.YRel * mouseSensitivity
		}
		viewRotation = mgl64.Vec2{xRel, yRel}
		controlVector = apputils.GetControlVector(keyboardInput)
	}

	forwardVector := g.camera.Rotation.Rotate(mgl64.Vec3{0, 0, -1})
	upVector := g.camera.Rotation.Rotate(mgl64.Vec3{0, 1, 0})
	// there's probably away to get the right vector directly rather than going crossing the up vector :D
	rightVector := forwardVector.Cross(upVector)

	// calculate the quaternion for the delta in rotation
	deltaRotationX := mgl64.QuatRotate(viewRotation[1], rightVector)         // pitch
	deltaRotationY := mgl64.QuatRotate(viewRotation[0], mgl64.Vec3{0, 1, 0}) // yaw
	deltaRotation := deltaRotationY.Mul(deltaRotationX)

	newRotation := deltaRotation.Mul(g.camera.Rotation)

	// don't let the camera go upside down
	if newRotation.Rotate(mgl64.Vec3{0, 1, 0})[1] < 0 {
		newRotation = g.camera.Rotation
	}

	g.camera.Rotation = newRotation

	// keyboardInput := frameInput.KeyboardInput
	// controlVector := getControlVector(keyboardInput)
	if !g.MouseCaptured() {
		controlVector = mgl64.Vec3{}
	}

	movementVector := rightVector.Mul(controlVector[0]).Add(mgl64.Vec3{0, 1, 0}.Mul(controlVector[1])).Add(forwardVector.Mul(controlVector[2]))

	if !movementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
		if g.camera.LastFrameMovementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
			// this is the starting speed that the camera accelerates from
			g.camera.Speed = settings.CameraSpeed * 0.3
		} else {
			// TODO(kevin) parameterize how slowly we accelerate based on how long we want to drift for
			g.camera.Speed *= 1.03
			if g.camera.Speed > settings.CameraSpeed {
				g.camera.Speed = settings.CameraSpeed
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

	if _, ok := keyboardInput[input.KeyboardKeyUp]; ok {
		g.camera.Position = g.camera.Position.Add(forwardVector.Mul(settings.CameraSlowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}
	if _, ok := keyboardInput[input.KeyboardKeyDown]; ok {
		g.camera.Position = g.camera.Position.Add(forwardVector.Mul(-settings.CameraSlowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}
	if _, ok := keyboardInput[input.KeyboardKeyLeft]; ok {
		g.camera.Position = g.camera.Position.Add(rightVector.Mul(-settings.CameraSlowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}
	if _, ok := keyboardInput[input.KeyboardKeyRight]; ok {
		g.camera.Position = g.camera.Position.Add(rightVector.Mul(settings.CameraSlowSpeed).Mul(float64(delta.Milliseconds()) / 1000))
	}

	g.camera.LastFrameMovementVector = movementVector
}

func (g *Client) handleGizmos(frameInput input.Input) {
	mouseInput := frameInput.MouseInput

	// set gizmo mode
	if g.SelectedEntity() != nil {
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
	e := g.SelectedEntity()

	if e != nil {
		if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.TranslationGizmo, e, g.runtimeConfig.SnapSize)
			if delta != nil {
				if e.Parent != nil {
					// the computed position is in world space but e.LocalPosition is in local space
					// to compute the new local space position we need to do conversions

					// compute the full transformation matrix, excluding local transformations
					// i.e. local transformations should not affect how the gizmo affects the entity
					transformMatrix := entity.ComputeParentAndJointTransformMatrix(e)

					// take the new world position and convert it to local space
					worldPosition := e.Position().Add(*delta)
					newPositionInLocalSpace := transformMatrix.Inv().Mul4x1(worldPosition.Vec4(1)).Vec3()

					entity.SetLocalPosition(e, newPositionInLocalSpace)
				} else {
					entity.SetLocalPosition(e, e.LocalPosition.Add(*delta))
				}
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewPositionEdit(gizmo.TranslationGizmo.ActivationPosition, e.GetLocalPosition(), e),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.TranslationGizmo.ActivationPosition = e.GetLocalPosition()
				gizmo.TranslationGizmo.LastSnapVector = e.GetLocalPosition()
			}
			gizmoHovered = gizmo.TranslationGizmo.HoveredEntityID != -1
		} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.RotationGizmo, e, g.runtimeConfig.SnapSize)
			if delta != nil {
				var magnitude float64 = 0

				if math.Abs(delta.X()) >= math.Abs(delta.Y()) {
					magnitude = delta.X()
				} else {
					magnitude = delta.Y()
				}
				magnitude *= math.Pi / float64(g.runtimeConfig.RotationSensitivity)

				forwardVector := g.camera.Rotation.Rotate(mgl64.Vec3{0, 0, -1})

				var newRotationAdjustment mgl64.Quat
				if gizmo.RotationGizmo.HoveredEntityID == gizmo.GizmoXDistancePickingID {
					if forwardVector.Dot(mgl64.Vec3{0, 0, -1}) > 0 {
						newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, 0, -1})
					} else {
						newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, 0, 1})
					}
				} else if gizmo.RotationGizmo.HoveredEntityID == gizmo.GizmoYDistancePickingID {
					if forwardVector.Dot(mgl64.Vec3{1, 0, 0}) > 0 {
						newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{1, 0, 0})
					} else {
						newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{-1, 0, 0})
					}
				} else if gizmo.RotationGizmo.HoveredEntityID == gizmo.GizmoZDistancePickingID {
					if forwardVector.Dot(mgl64.Vec3{0, -1, 0}) > 0 {
						newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, -1, 0})
					} else {
						newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, 1, 0})
					}
				} else {
					panic("wat")
				}

				if e.Parent != nil {
					transformMatrix := entity.ComputeParentAndJointTransformMatrix(e)
					worldToLocalMatrix := transformMatrix.Inv()
					_, r, _ := utils.DecomposeF64(worldToLocalMatrix)
					computedRotation := r.Mul(newRotationAdjustment)
					e.SetLocalRotation(computedRotation.Mul(e.GetLocalRotation()))
				} else {
					e.SetLocalRotation(newRotationAdjustment.Mul(e.GetLocalRotation()))
				}
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewRotationEdit(gizmo.TranslationGizmo.ActivationRotation, e.GetLocalRotation(), e),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.RotationGizmo.ActivationRotation = e.GetLocalRotation()
				// gizmo.TranslationGizmo.LastSnapVector = mgl64.Vec3{}
			}
			gizmoHovered = gizmo.RotationGizmo.HoveredEntityID != -1
		} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.ScaleGizmo, e, g.runtimeConfig.SnapSize)
			if delta != nil {
				magnitude := settings.ScaleSensitivity
				if gizmo.ScaleGizmo.HoveredEntityID == gizmo.GizmoAllAxisPickingID {
					magnitude = settings.ScaleAllAxisSensitivity
				}
				scale := e.Scale()
				entity.SetScale(e, scale.Add(delta.Mul(magnitude)))
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewScaleEdit(gizmo.ScaleGizmo.ActivationScale, e.Scale(), e),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.ScaleGizmo.ActivationScale = e.Scale()
			}
			gizmoHovered = gizmo.ScaleGizmo.HoveredEntityID != -1
		}
	}

	if !gizmoHovered && g.renderSystem.GameWindowHovered() && mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
		entityID := g.renderSystem.TryHoverEntity()
		if entityID == nil || g.world.GetEntityByID(*entityID) == nil {
			g.SelectEntity(nil)
			gizmo.CurrentGizmoMode = gizmo.GizmoModeNone
		} else {
			clickedEntity := g.world.GetEntityByID(*entityID)
			currentSelection := g.SelectedEntity()

			if currentSelection != nil && currentSelection.ID != clickedEntity.ID {
				gizmo.CurrentGizmoMode = gizmo.GizmoModeNone
			}

			g.SelectEntity(clickedEntity)
		}
	}
}

func (g *Client) updateGizmo(frameInput input.Input, targetGizmo *gizmo.Gizmo, e *entity.Entity, snapSize float64) (*mgl64.Vec3, gizmo.GizmoEvent) {
	mouseInput := frameInput.MouseInput
	colorPickingID := g.renderSystem.HoveredEntityID()

	gameWindowWidth, gameWindowHeight := g.renderSystem.SceneSize()
	nearPlanePos := g.mousePosToNearPlane(mouseInput.Position, gameWindowWidth, gameWindowHeight)

	cameraViewDir := g.camera.Rotation.Rotate(mgl64.Vec3{0, 0, -1})
	delta, gizmoEvent := gizmo.CalculateGizmoDelta(targetGizmo, frameInput, cameraViewDir, e.Position(), g.camera.Position, nearPlanePos, colorPickingID, snapSize)
	return delta, gizmoEvent
}
