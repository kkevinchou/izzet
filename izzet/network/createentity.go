package network

import "github.com/go-gl/mathgl/mgl64"

type CreateEntityMessage struct {
	Position    mgl64.Vec3
	Orientation mgl64.Quat
	Scale       mgl64.Vec3
	OwnerID     int
}
