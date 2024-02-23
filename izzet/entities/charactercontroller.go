package entities

import "github.com/go-gl/mathgl/mgl64"

type CharacterControllerComponent struct {
	ControlVector       mgl64.Vec3
	Speed               float64
	FlySpeed            float64
	WebVector           mgl64.Vec3
	PersistentWebVector mgl64.Vec3
}
