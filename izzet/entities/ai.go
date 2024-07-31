package entities

import "github.com/go-gl/mathgl/mgl64"

type AIComponent struct {
	PatrolConfig   *PatrolConfig
	RotationConfig *RotationConfig
	TargetConfig   *TargetConfig
	PathfindConfig *PathfindConfig
	Speed          float64
}

type TargetConfig struct {
	Direction mgl64.Vec3
}

type PatrolConfig struct {
	Points []mgl64.Vec3
	Index  int
}

type PathfindConfig struct {
	Target mgl64.Vec3
	Path   []mgl64.Vec3
}

type RotationConfig struct {
	Quat mgl64.Quat
}
