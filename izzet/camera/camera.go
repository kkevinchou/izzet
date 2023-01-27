package camera

import "github.com/go-gl/mathgl/mgl64"

type Camera struct {
	Position    mgl64.Vec3
	Orientation mgl64.Quat
	Speed       float64

	Drift                   mgl64.Vec3
	LastFrameMovementVector mgl64.Vec3
}
