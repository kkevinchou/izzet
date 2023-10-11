package entities

import (
	"github.com/go-gl/mathgl/mgl64"
)

type CharacterControllerComponent struct {
	Orientation      mgl64.Quat
	ControlDirection mgl64.Vec3
	Speed            float64
}
