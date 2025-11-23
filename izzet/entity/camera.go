package entity

import "github.com/go-gl/mathgl/mgl64"

type CameraComponent struct {
	TargetPositionOffset mgl64.Vec3
	Target               *int
}
