package entities

import "github.com/go-gl/mathgl/mgl64"

type MovementComponent struct {
	PatrolConfig *PatrolConfig
	Speed        float64
}

type PatrolConfig struct {
	Points []mgl64.Vec3
	Index  int
}
