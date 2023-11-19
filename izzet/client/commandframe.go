package client

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/kkevinchou/kitolib/utils"
)

var (
	maxCameraSpeed float64 = 400 // units per second
	slowSpeed      float64 = 100 // units per second
)

// Systems Context

func (g *Client) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	frameInput := g.GetFrameInput()

	if frameInput.WindowEvent.Resized {
		w, h := g.window.GetSize()
		g.SetWindowSize(int(w), int(h))
		g.renderer.ReinitializeFrameBuffers()
	}

	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	if g.RuntimeConfig().EnableSpatialPartition {
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

	g.RuntimeConfig().CameraPosition = g.camera.Position
	g.RuntimeConfig().CameraRotation = g.camera.Rotation
}

func (g *Client) handleSpatialPartition() {
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

func (g *Client) handleInputCommands(frameInput input.Input) {
	for _, cmd := range frameInput.Commands {
		if c, ok := cmd.(input.FileDropCommand); ok {
			fmt.Println("received drop file command", c.File)
		} else if _, ok := cmd.(input.QuitCommand); ok {
			g.Shutdown()
		}
	}

	mouseInput := frameInput.MouseInput
	keyboardInput := frameInput.KeyboardInput

	// shutdown
	if event, ok := keyboardInput[input.KeyboardKeyEscape]; ok && event.Event == input.KeyboardEventUp {
		if g.AppMode() == app.AppModeEditor {
			g.Shutdown()
		} else if g.AppMode() == app.AppModePlay {
			g.DisconnectClient()
		}
	}

	if g.renderer.GameWindowHovered() {
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

	if _, ok := keyboardInput[input.KeyboardKeyF5]; ok {
		err := g.ConnectAndInitialize()
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

					serialization.InitDeserializedEntity(&newEntity, g.ModelLibrary())

					g.world.AddEntity(&newEntity)
					panels.SelectEntity(&newEntity)
				}
			}
		}
	}

	// navmesh - move highlight
	if event, ok := keyboardInput[input.KeyboardKeyI]; ok {
		if event.Event == input.KeyboardEventUp {
			g.RuntimeConfig().VoxelHighlightZ--
			g.ResetNavMeshVAO()
		}
	}
	if event, ok := keyboardInput[input.KeyboardKeyK]; ok {
		if event.Event == input.KeyboardEventUp {
			g.RuntimeConfig().VoxelHighlightZ++
			g.ResetNavMeshVAO()
		}
	}
	if event, ok := keyboardInput[input.KeyboardKeyJ]; ok {
		if event.Event == input.KeyboardEventUp {
			g.RuntimeConfig().VoxelHighlightX--
			g.ResetNavMeshVAO()
		}
	}
	if event, ok := keyboardInput[input.KeyboardKeyL]; ok {
		if event.Event == input.KeyboardEventUp {
			g.RuntimeConfig().VoxelHighlightX++
			g.ResetNavMeshVAO()
		}
	}
}

func (g *Client) editorCameraMovement(frameInput input.Input, delta time.Duration) {
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
		controlVector = app.GetControlVector(keyboardInput)
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

func (g *Client) handleGizmos(frameInput input.Input) {
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
	entity := panels.SelectedEntity()

	if entity != nil {
		if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
			snapSize := int(g.runtimeConfig.SnapSize)
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.TranslationGizmo, entity, snapSize)
			if delta != nil {
				if entity.Parent != nil {
					// the computed position is in world space but entity.LocalPosition is in local space
					// to compute the new local space position we need to do conversions

					// compute the full transformation matrix, excluding local transformations
					// i.e. local transformations should not affect how the gizmo affects the entity
					transformMatrix := entities.ComputeParentAndJointTransformMatrix(entity)

					// take the new world position and convert it to local space
					worldPosition := entity.Position().Add(*delta)
					newPositionInLocalSpace := transformMatrix.Inv().Mul4x1(worldPosition.Vec4(1)).Vec3()

					entities.SetLocalPosition(entity, newPositionInLocalSpace)
				} else {
					entities.SetLocalPosition(entity, entity.LocalPosition.Add(*delta))
				}
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewPositionEdit(gizmo.TranslationGizmo.ActivationPosition, entities.GetLocalPosition(entity), entity),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.TranslationGizmo.ActivationPosition = entities.GetLocalPosition(entity)
				gizmo.TranslationGizmo.LastSnapVector = entities.GetLocalPosition(entity)
			}
			gizmoHovered = gizmo.TranslationGizmo.HoveredEntityID != -1
		} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
			snapSize := int(g.runtimeConfig.RotationSnapSize)
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.RotationGizmo, entity, snapSize)
			if delta != nil {
				var magnitude float64 = 0

				if math.Abs(delta.X()) >= math.Abs(delta.Y()) {
					magnitude = delta.X()
				} else {
					magnitude = delta.Y()
				}
				magnitude *= math.Pi / float64(g.runtimeConfig.RotationSensitivity)

				var newRotationAdjustment mgl64.Quat
				if gizmo.RotationGizmo.HoveredEntityID == gizmo.GizmoXDistancePickingID {
					newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, 0, 1})
				} else if gizmo.RotationGizmo.HoveredEntityID == gizmo.GizmoYDistancePickingID {
					newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{1, 0, 0})
				} else if gizmo.RotationGizmo.HoveredEntityID == gizmo.GizmoZDistancePickingID {
					newRotationAdjustment = mgl64.QuatRotate(magnitude, mgl64.Vec3{0, 1, 0})
				} else {
					panic("wat")
				}

				if entity.Parent != nil {
					transformMatrix := entities.ComputeParentAndJointTransformMatrix(entity)
					worldToLocalMatrix := transformMatrix.Inv()
					_, r, _ := utils.DecomposeF64(worldToLocalMatrix)
					computedRotation := r.Mul(newRotationAdjustment)
					entities.SetLocalRotation(entity, computedRotation.Mul(entities.GetLocalRotation(entity)))
				} else {
					entities.SetLocalRotation(entity, newRotationAdjustment.Mul(entities.GetLocalRotation(entity)))
				}
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewRotationEdit(gizmo.TranslationGizmo.ActivationRotation, entities.GetLocalRotation(entity), entity),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.RotationGizmo.ActivationRotation = entities.GetLocalRotation(entity)
				// gizmo.TranslationGizmo.LastSnapVector = mgl64.Vec3{}
			}
			gizmoHovered = gizmo.RotationGizmo.HoveredEntityID != -1
		} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
			snapSize := int(g.runtimeConfig.SnapSize)
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.ScaleGizmo, entity, snapSize)
			if delta != nil {
				magnitude := 0.05
				if gizmo.ScaleGizmo.HoveredEntityID == gizmo.GizmoAllAxisPickingID {
					magnitude = 0.005
				}
				scale := entities.GetLocalScale(entity)
				entities.SetScale(entity, scale.Add(delta.Mul(magnitude)))
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewScaleEdit(gizmo.ScaleGizmo.ActivationScale, entities.GetLocalScale(entity), entity),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.ScaleGizmo.ActivationScale = entities.GetLocalScale(entity)
			}
			gizmoHovered = gizmo.ScaleGizmo.HoveredEntityID != -1
		}
	}

	if !gizmoHovered && g.renderer.GameWindowHovered() && mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
		convertedMousePosition := mgl64.Vec2{mouseInput.Position.X(), mouseInput.Position.Y() + float64(g.renderer.CalculateFooterSize())}
		entityID := g.renderer.GetEntityByPixelPosition(convertedMousePosition)
		if entityID == nil || g.world.GetEntityByID(*entityID) == nil {
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

func (g *Client) updateGizmo(frameInput input.Input, targetGizmo *gizmo.Gizmo, entity *entities.Entity, snapSize int) (*mgl64.Vec3, gizmo.GizmoEvent) {
	mouseInput := frameInput.MouseInput
	convertedMousePosition := mgl64.Vec2{mouseInput.Position.X(), mouseInput.Position.Y() + float64(g.renderer.CalculateFooterSize())}
	colorPickingID := g.renderer.GetEntityByPixelPosition(convertedMousePosition)
	nearPlanePos := g.mousePosToNearPlane(convertedMousePosition, g.width, g.height)

	delta, gizmoEvent := gizmo.CalculateGizmoDelta(targetGizmo, frameInput, entity.Position(), g.camera.Position, nearPlanePos, colorPickingID, snapSize)
	return delta, gizmoEvent
}
