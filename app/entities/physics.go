package entities

import "github.com/go-gl/mathgl/mgl64"

type PhysicsComponent struct {
	Velocity         mgl64.Vec3
	Grounded         bool
	GravityEnabled   bool
	OrientOnVelocity bool
}
