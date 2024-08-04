package entities

import "github.com/go-gl/mathgl/mgl32"

type CameraComponent struct {
	TargetPositionOffset mgl32.Vec3
	Target               *int
}
