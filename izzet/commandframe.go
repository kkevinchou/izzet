package izzet

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/kitolib/input"
)

func (g *Izzet) handleResize() {
	w, h := g.window.GetSize()
	g.aspectRatio = float64(w) / float64(h)
	g.fovY = mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovx)/2)/g.aspectRatio))
}

func (g *Izzet) runCommandFrame(frameInput input.Input, delta time.Duration) {
	g.handleResize()

	gizmo.T.Reset()

	for _, entity := range g.entities {
		if entity.AnimationPlayer != nil {
			entity.AnimationPlayer.Update(delta)
		}
	}

	keyboardInput := frameInput.KeyboardInput
	if _, ok := keyboardInput[input.KeyboardKeyEscape]; ok {
		// move this into a system maybe
		g.Shutdown()
	}

	var xRel, yRel float64

	mouseInput := frameInput.MouseInput

	var mouseSensitivity float64 = 0.003
	if mouseInput.Buttons[1] && !mouseInput.MouseMotionEvent.IsZero() {
		xRel += -mouseInput.MouseMotionEvent.XRel * mouseSensitivity
		yRel += -mouseInput.MouseMotionEvent.YRel * mouseSensitivity
	}

	// convert gizmo axes to world space
	// line point distance check
	// if within threshold, adjust
	if panels.SelectedEntity != nil {
		points := []mgl64.Vec3{panels.SelectedEntity.Position, panels.SelectedEntity.Position.Add(mgl64.Vec3{0, 20, 0})}
		// points := []mgl64.Vec3{panels.SelectedEntity.Position}
		screenSpacePoints := []mgl64.Vec2{}

		w, h := g.window.GetSize()
		for _, p := range points {
			newP := g.viewerContext.ProjectionMatrix.Mul4(g.viewerContext.InverseViewMatrix).Mul4x1(p.Vec4(1))
			ndcP := newP.Mul(1.0 / newP.W()).Vec3()

			screenP := mgl64.Vec2{(ndcP.X()/2 + 0.5) * float64(w), (1 - (ndcP.Y()/2 + 0.5)) * float64(h)}
			screenSpacePoints = append(screenSpacePoints, screenP)
		}

		if mouseInput.Buttons[0] && !mouseInput.MouseMotionEvent.IsZero() {
			// closestPoint := checks.ClosestPointOnLineToPoint(screenSpacePoints[0].Vec3(0), screenSpacePoints[1].Vec3(0), mouseInput.Position.Vec3(0))
			closestPoint := screenSpacePoints[0].Vec3(0)
			if mouseInput.Position.Vec3(0).Sub(closestPoint).Len() < 100 {
				gizmo.T.Move(gizmo.AxisTypeY, -mouseInput.MouseMotionEvent.YRel)
				// gizmo.T.Move(gizmo.AxisTypeX, mouseInput.MouseMotionEvent.XRel)
			}
		}
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

	cameraSpeed := 100
	controlVector := getControlVector(keyboardInput)
	movementVector := rightVector.Mul(controlVector[0]).Add(mgl64.Vec3{0, 1, 0}.Mul(controlVector[1])).Add(forwardVector.Mul(controlVector[2]))
	movementDelta := movementVector.Mul(float64(cameraSpeed) / float64(delta.Milliseconds()))

	g.camera.Position = g.camera.Position.Add(movementDelta)
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
