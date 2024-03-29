package entities

import "github.com/go-gl/mathgl/mgl64"

type AIComponent struct {
	PatrolConfig   *PatrolConfig
	RotationConfig *RotationConfig
	TargetConfig   *TargetConfig
	Speed          float64
}

type TargetConfig struct {
	Direction mgl64.Vec3
}

type PatrolConfig struct {
	Points []mgl64.Vec3
	Index  int
}

type RotationConfig struct {
	Quat mgl64.Quat
}
