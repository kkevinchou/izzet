package entities

import "github.com/go-gl/mathgl/mgl64"

type CameraComponent struct {
	PositionOffset mgl64.Vec3
	Target         *int
}
