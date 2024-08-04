package entities

import "github.com/go-gl/mathgl/mgl32"

type PhysicsComponent struct {
	Velocity         mgl32.Vec3
	Grounded         bool
	GravityEnabled   bool
	RotateOnVelocity bool
}
