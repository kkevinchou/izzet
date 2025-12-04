package entity

import "github.com/go-gl/mathgl/mgl64"

type CharacterControllerComponent struct {
	ControlVector       mgl64.Vec3
	FlySpeed            float64
	WebVector           mgl64.Vec3
	PersistentWebVector mgl64.Vec3
}
