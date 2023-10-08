package entities

import "github.com/go-gl/mathgl/mgl64"

type PhysicsComponent struct {
	Velocity mgl64.Vec3

	// static entities do not actively initiate collisions, the expectation is that if there's a collision
	// the other entity is the only entity that will move and thus resolve the collision
	Static bool
}
