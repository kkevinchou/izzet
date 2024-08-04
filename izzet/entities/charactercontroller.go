package entities

import "github.com/go-gl/mathgl/mgl32"

type CharacterControllerComponent struct {
	ControlVector       mgl32.Vec3
	Speed               float32
	FlySpeed            float32
	WebVector           mgl32.Vec3
	PersistentWebVector mgl32.Vec3
}
