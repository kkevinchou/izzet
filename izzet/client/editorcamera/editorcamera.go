package editorcamera

import "github.com/go-gl/mathgl/mgl32"

type Camera struct {
	Position mgl32.Vec3
	Rotation mgl32.Quat
	Speed    float32

	Drift                   mgl32.Vec3
	LastFrameMovementVector mgl32.Vec3
}
