package client

import (
	"fmt"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/checks"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/navmesh"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/client/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/render"
	"github.com/kkevinchou/izzet/izzet/render/panels"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

// Systems Context

func (g *Client) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	frameInput := g.GetFrameInput()

	if g.platform.Resized() {
		w, h := g.window.GetSize()
		g.SetWindowSize(w, h)
		g.renderSystem.ReinitializeFrameBuffers()
	}

	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	g.handleSpatialPartition()

	if g.AppMode() == types.AppModePlay {
		for _, s := range g.playModeSystems {
			s.Update(delta, g.world)
		}
	} else if g.AppMode() == types.AppModeEditor {
		for _, s := range g.editorModeSystems {
			s.Update(delta, g.world)
		}
	}

	g.handleInputCommands(frameInput)

	if g.AppMode() == types.AppModeEditor {
		g.handleEditorInputCommands(frameInput)
		g.editorCameraMovement(frameInput, delta)
		g.handleGizmos(frameInput)
	} else if g.AppMode() == types.AppModePlay {
		g.handlePlayInputCommands(frameInput)
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

func (g *Client) handleEditorInputCommands(frameInput input.Input) {
	// mouseInput := frameInput.MouseInput
	keyboardInput := frameInput.KeyboardInput

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
				if entity := g.SelectedEntity(); entity != nil {
					var err error
					copiedEntity, err = serialization.SerializeEntity(entity)
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
					id := entities.GetNextIDAndAdvance()
					e.ID = id
					g.world.AddEntity(e)
					g.SelectEntity(e)
				}
			}
		}
	}
	// set pathfinding start and goal

	if event, ok := keyboardInput[input.KeyboardKeyN]; ok {
		if event.Event == input.KeyboardEventUp {
			if g.navMesh != nil {
				mousePosition := frameInput.MouseInput.Position
				width, height := g.renderSystem.SceneSize()
				ctx := g.renderSystem.CameraViewerContext()

				xNDC := (mousePosition.X()/float64(width) - 0.5) * 2

				menuBarSize := float64(render.CalculateMenuBarHeight())
				yNDC := ((float64(height)-mousePosition.Y()+menuBarSize)/float64(height) - 0.5) * 2

				nearPlanePosition := rutils.NDCToWorldPosition(ctx, mgl64.Vec3{xNDC, yNDC, -float64(g.RuntimeConfig().Near)})
				point, success := g.intersectRayWithEntities(g.GetEditorCameraPosition(), nearPlanePosition.Sub(g.GetEditorCameraPosition()).Normalize())

				if success {
					c := navmesh.CompileNavMesh(g.navMesh)
					pt, p, success := navmesh.FindNearestPolygon(c.Tiles[0], point)
					if success {
						g.runtimeConfig.NavigationMeshStart = int32(p)
						g.runtimeConfig.NavigationMeshStartPoint = pt
					}
				}
			}
		}
	}
	if event, ok := keyboardInput[input.KeyboardKeyM]; ok {
		if event.Event == input.KeyboardEventUp {
			mousePosition := frameInput.MouseInput.Position
			width, height := g.renderSystem.SceneSize()
			ctx := g.renderSystem.CameraViewerContext()

			xNDC := (mousePosition.X()/float64(width) - 0.5) * 2

			menuBarSize := float64(render.CalculateMenuBarHeight())
			yNDC := ((float64(height)-mousePosition.Y()+menuBarSize)/float64(height) - 0.5) * 2

			nearPlanePosition := rutils.NDCToWorldPosition(ctx, mgl64.Vec3{xNDC, yNDC, -float64(g.RuntimeConfig().Near)})
			point, success := g.intersectRayWithEntities(g.GetEditorCameraPosition(), nearPlanePosition.Sub(g.GetEditorCameraPosition()).Normalize())

			if success && g.navMesh != nil {
				c := navmesh.CompileNavMesh(g.navMesh)
				pt, p, success := navmesh.FindNearestPolygon(c.Tiles[0], point)
				if success {
					g.runtimeConfig.NavigationMeshGoal = int32(p)
					g.runtimeConfig.NavigationMeshGoalPoint = pt
				}
			}
		}
	}
}

func (g *Client) handlePlayInputCommands(frameInput input.Input) {
	keyboardInput := frameInput.KeyboardInput

	// set ai pathfinding target
	if event, ok := keyboardInput[input.KeyboardKeyN]; ok {
		if event.Event == input.KeyboardEventUp {
			mousePosition := frameInput.MouseInput.Position
			width, height := g.renderSystem.SceneSize()
			ctx := g.renderSystem.CameraViewerContext()

			xNDC := (mousePosition.X()/float64(width) - 0.5) * 2

			menuBarSize := float64(render.CalculateMenuBarHeight())
			yNDC := ((float64(height)-mousePosition.Y()+menuBarSize)/float64(height) - 0.5) * 2

			nearPlanePosition := rutils.NDCToWorldPosition(ctx, mgl64.Vec3{xNDC, yNDC, -float64(g.RuntimeConfig().Near)})
			camera := g.GetPlayerCamera()
			position := camera.Position()
			point, success := g.intersectRayWithEntities(position, nearPlanePosition.Sub(position).Normalize())

			if success {
				rpcMessage := network.RPCMessage{
					Pathfind: &network.Pathfind{Goal: point},
				}
				g.Client().Send(rpcMessage, g.CommandFrame())
			}
		}
	}

	if event, ok := keyboardInput[input.KeyboardKeyJ]; ok {
		if event.Event == input.KeyboardEventUp {
			rpcMessage := network.RPCMessage{
				CreateEntity: &network.CreateEntityRPC{EntityType: string(panels.SelectedCreateEntityComboOption), Patrol: true},
			}
			g.Client().Send(rpcMessage, g.CommandFrame())
		}
	}

	if event, ok := keyboardInput[input.KeyboardKeyK]; ok {
		if event.Event == input.KeyboardEventUp {
			rpcMessage := network.RPCMessage{
				CreateEntity: &network.CreateEntityRPC{EntityType: string(panels.SelectedCreateEntityComboOption)},
			}
			g.Client().Send(rpcMessage, g.CommandFrame())
		}
	}
}

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

	if event, ok := keyboardInput[input.KeyboardKeyF1]; ok && event.Event == input.KeyboardEventUp {
		g.ConfigureUI(!g.runtimeConfig.UIEnabled)
	}

	// shutdown
	if event, ok := keyboardInput[input.KeyboardKeyEscape]; ok && event.Event == input.KeyboardEventUp {
		if g.AppMode() == types.AppModeEditor {
			g.Shutdown()
		} else if g.AppMode() == types.AppModePlay {
			g.DisconnectClient()
		}
	}

	if g.renderSystem.GameWindowHovered() {
		if mouseInput.MouseButtonEvent[1] == input.MouseButtonEventDown {
			g.relativeMouseActive = true
			g.relativeMouseOrigin[0] = int32(mouseInput.Position[0])
			g.relativeMouseOrigin[1] = int32(mouseInput.Position[1])
			g.platform.SetRelativeMouse(true)
		}
	}

	// we should continue to keep the mouse at the origin, regardless of
	// whether we're hoving the game window or not
	if g.relativeMouseActive {
		g.platform.MoveMouse(g.relativeMouseOrigin[0], g.relativeMouseOrigin[1])

		if mouseInput.MouseButtonEvent[1] == input.MouseButtonEventUp {
			g.relativeMouseActive = false
			g.platform.SetRelativeMouse(false)
			g.platform.MoveMouse(g.relativeMouseOrigin[0], g.relativeMouseOrigin[1])
		}
	}
}

func (g *Client) intersectRayWithEntities(position, dir mgl64.Vec3) (mgl64.Vec3, bool) {
	var hit bool
	var hitPoint mgl64.Vec3

	minDistSq := math.MaxFloat64

	// TODO: this should ray cast against navmesh polys, not entity geometry
	for _, entity := range g.world.Entities() {
		if entity.Collider == nil || entity.Collider.TriMeshCollider == nil {
			continue
		}

		ray := collider.Ray{Origin: position, Direction: dir}
		transformMatrix := entities.WorldTransform(entity)
		collider := entity.Collider.TriMeshCollider.Transform(transformMatrix)

		point, success := checks.IntersectRayTriMesh(ray, collider)
		if !success {
			continue
		}

		hit = true

		distSq := position.Sub(point).LenSqr()
		if distSq < minDistSq {
			minDistSq = distSq
			hitPoint = point
		}

	}

	return hitPoint, hit
}

func (g *Client) editorCameraMovement(frameInput input.Input, delta time.Duration) {
	mouseInput := frameInput.MouseInput
	keyboardInput := frameInput.KeyboardInput

	var viewRotation mgl64.Vec2
	var controlVector mgl64.Vec3
	if g.relativeMouseActive {
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
	if !frameInput.MouseInput.MouseButtonState[1] {
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
	entity := g.SelectedEntity()

	if entity != nil {
		if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.TranslationGizmo, entity, g.runtimeConfig.SnapSize)
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
					edithistory.NewPositionEdit(gizmo.TranslationGizmo.ActivationPosition, entity.GetLocalPosition(), entity),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.TranslationGizmo.ActivationPosition = entity.GetLocalPosition()
				gizmo.TranslationGizmo.LastSnapVector = entity.GetLocalPosition()
			}
			gizmoHovered = gizmo.TranslationGizmo.HoveredEntityID != -1
		} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.RotationGizmo, entity, g.runtimeConfig.SnapSize)
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

				if entity.Parent != nil {
					transformMatrix := entities.ComputeParentAndJointTransformMatrix(entity)
					worldToLocalMatrix := transformMatrix.Inv()
					_, r, _ := utils.DecomposeF64(worldToLocalMatrix)
					computedRotation := r.Mul(newRotationAdjustment)
					entity.SetLocalRotation(computedRotation.Mul(entity.GetLocalRotation()))
				} else {
					entity.SetLocalRotation(newRotationAdjustment.Mul(entity.GetLocalRotation()))
				}
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewRotationEdit(gizmo.TranslationGizmo.ActivationRotation, entity.GetLocalRotation(), entity),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.RotationGizmo.ActivationRotation = entity.GetLocalRotation()
				// gizmo.TranslationGizmo.LastSnapVector = mgl64.Vec3{}
			}
			gizmoHovered = gizmo.RotationGizmo.HoveredEntityID != -1
		} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
			delta, gizmoEvent := g.updateGizmo(frameInput, gizmo.ScaleGizmo, entity, g.runtimeConfig.SnapSize)
			if delta != nil {
				magnitude := settings.ScaleSensitivity
				if gizmo.ScaleGizmo.HoveredEntityID == gizmo.GizmoAllAxisPickingID {
					magnitude = settings.ScaleAllAxisSensitivity
				}
				scale := entity.Scale()
				entities.SetScale(entity, scale.Add(delta.Mul(magnitude)))
			} else if gizmoEvent == gizmo.GizmoEventCompleted {
				g.AppendEdit(
					edithistory.NewScaleEdit(gizmo.ScaleGizmo.ActivationScale, entity.Scale(), entity),
				)
			}
			if gizmoEvent == gizmo.GizmoEventActivated {
				gizmo.ScaleGizmo.ActivationScale = entity.Scale()
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

func (g *Client) updateGizmo(frameInput input.Input, targetGizmo *gizmo.Gizmo, entity *entities.Entity, snapSize float64) (*mgl64.Vec3, gizmo.GizmoEvent) {
	mouseInput := frameInput.MouseInput
	colorPickingID := g.renderSystem.HoveredEntityID()

	gameWindowWidth, gameWindowHeight := g.renderSystem.SceneSize()
	nearPlanePos := g.mousePosToNearPlane(mouseInput.Position, gameWindowWidth, gameWindowHeight)

	cameraViewDir := g.camera.Rotation.Rotate(mgl64.Vec3{0, 0, -1})
	delta, gizmoEvent := gizmo.CalculateGizmoDelta(targetGizmo, frameInput, cameraViewDir, entity.Position(), g.camera.Position, nearPlanePos, colorPickingID, snapSize)
	return delta, gizmoEvent
}
