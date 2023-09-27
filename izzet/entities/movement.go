package entities

import "github.com/go-gl/mathgl/mgl64"

type MovementComponent struct {
	PatrolConfig   *PatrolConfig
	RotationConfig *RotationConfig
	Speed          float64
}

type PatrolConfig struct {
	Points []mgl64.Vec3
	Index  int
}

type RotationConfig struct {
	Quat mgl64.Quat
}
