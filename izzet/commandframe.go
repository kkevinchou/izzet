package izzet

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/collision/checks"
	"github.com/kkevinchou/kitolib/input"
)

var (
	maxUInt32 uint32 = ^uint32(0)

	// we shift 8 bits since 8 bits are reserved for the alpha channel
	// the max id is used to indicate no entity was selected
	emptyColorPickingID uint32 = maxUInt32 >> 8
)

func (g *Izzet) handleResize() {
	w, h := g.window.GetSize()
	g.aspectRatio = float64(w) / float64(h)
	g.fovY = mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovx)/2)/g.aspectRatio))
}

func (g *Izzet) mousePosToNearPlane(mouseInput input.MouseInput) mgl64.Vec3 {
	w, h := g.Window().GetSize()
	x := mouseInput.Position.X()
	y := mouseInput.Position.Y()

	// -1 for the near plane
	ndcP := mgl64.Vec4{((x / float64(w)) - 0.5) * 2, ((y / float64(h)) - 0.5) * -2, -1, 1}
	nearPlanePos := g.viewerContext.InverseViewMatrix.Inv().Mul4(g.viewerContext.ProjectionMatrix.Inv()).Mul4x1(ndcP)
	nearPlanePos = nearPlanePos.Mul(1.0 / nearPlanePos.W())

	return nearPlanePos.Vec3()
}

func (g *Izzet) runCommandFrame(frameInput input.Input, delta time.Duration) {
	g.handleResize()

	for _, entity := range g.Entities() {
		if entity.AnimationPlayer != nil {
			entity.AnimationPlayer.Update(delta)
		}
	}

	keyboardInput := frameInput.KeyboardInput
	if _, ok := keyboardInput[input.KeyboardKeyEscape]; ok {
		// move this into a system maybe
		g.Shutdown()
	}

	g.gizmo(frameInput)
	g.cameraMovement(frameInput, delta)
	g.entitySelect(frameInput, delta)
}

func (g *Izzet) entitySelect(frameInput input.Input, delta time.Duration) {
	// gizmo interactions supercede entity selection
	if gizmo.T.Active {
		return
	}

	mouseInput := frameInput.MouseInput

	gl.BindFramebuffer(gl.FRAMEBUFFER, g.colorPickingFB)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		data := make([]byte, 4)
		_, h := g.window.GetSize()
		gl.ReadPixels(int32(mouseInput.Position[0]), int32(h)-int32(mouseInput.Position[1]), 1, 1, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(data))

		// NOTE(kevin) actually not sure why, but this works
		// i would've expected to need to multiply by 255, but apparently it's handled somehow
		uintID := binary.LittleEndian.Uint32(data)

		if uintID != emptyColorPickingID {
			id := int(uintID)
			for i, e := range g.Entities() {
				if e.ID == id {
					panels.HierarchySelection = (1 << i)
				}
			}
		} else {
			panels.HierarchySelection = 0
		}
	}
}

func (g *Izzet) cameraMovement(frameInput input.Input, delta time.Duration) {
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
			g.camera.Speed = 3
		} else {
			// TODO(kevin) parameterize how slowly we accelerate based on how long we want to drift for
			g.camera.Speed *= 1.1
			if g.camera.Speed > 18 {
				g.camera.Speed = 18
			}
		}
	}

	movementDelta := movementVector.Mul(float64(g.camera.Speed) / float64(delta.Milliseconds()))

	if movementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
		// start drifting if we were moving last frame but not the current one
		if !g.camera.LastFrameMovementVector.ApproxEqual(mgl64.Vec3{0, 0, 0}) {
			g.camera.Drift = g.camera.LastFrameMovementVector.Mul(float64(g.camera.Speed) / float64(delta.Milliseconds()))
		} else {
			// TODO(kevin) parameterize how slowly we decay based on how long we want to drift for
			g.camera.Drift = g.camera.Drift.Mul(0.92)
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

func (g *Izzet) gizmo(frameInput input.Input) {
	mouseInput := frameInput.MouseInput

	if panels.SelectedEntity != nil {
		nearPlanePos := g.mousePosToNearPlane(mouseInput)
		position := panels.SelectedEntity.Position

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

				if minDist == nil || length < float64(*minDist) {
					minAxis = axis
					minDist = &length
					motionPivot = b
					closestAxisIndex = i
				}
			}
		}

		if minDist != nil {
			if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventDown {
				gizmo.T.Active = true
				gizmo.T.TranslationDir = minAxis
				gizmo.T.MotionPivot = motionPivot.Sub(position)
				gizmo.T.HoverIndex = closestAxisIndex
			}

			if !gizmo.T.Active {
				gizmo.T.HoverIndex = closestAxisIndex
			}
		} else {
			if !gizmo.T.Active {
				gizmo.T.HoverIndex = -1
			}
		}

		if !mouseInput.MouseMotionEvent.IsZero() {
			if _, b, nonParallel := checks.ClosestPointsInfiniteLines(g.camera.Position, nearPlanePos, position, position.Add(gizmo.T.TranslationDir)); nonParallel {
				if gizmo.T.Active && mouseInput.Buttons[0] {
					newPosition := b.Sub(gizmo.T.MotionPivot)
					newPosition[0] = float64(int(newPosition[0]))
					newPosition[1] = float64(int(newPosition[1]))
					newPosition[2] = float64(int(newPosition[2]))
					panels.SelectedEntity.Position = newPosition
				}
			}
		}

		if mouseInput.MouseButtonEvent[0] == input.MouseButtonEventUp {
			gizmo.T.Active = false
			gizmo.T.HoverIndex = closestAxisIndex
		}
	}
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
