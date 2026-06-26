package serversystem

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/system"
)

type PhysicsSystem struct{}

func NewPhysicsSystem(_ App) *PhysicsSystem {
	return &PhysicsSystem{}
}

func (s *PhysicsSystem) Name() string {
	return "PhysicsSystem"
}

func (s *PhysicsSystem) Update(delta time.Duration, world system.GameWorld) {
	physicsWorld := world.PhysicsWorld()
	physicsWorld.Step(delta)

	for _, e := range world.Entities() {
		if e.Physics == nil || e.Physics.BodyID == 0 {
			continue
		}

		body, ok := physicsWorld.Body(e.Physics.BodyID)
		if !ok {
			e.Physics.BodyID = 0
			continue
		}

		transform := body.Transform()
		entity.SetLocalPosition(e, transform.Position)
		e.SetLocalRotation(transform.Rotation)
		e.Physics.Velocity = body.LinearVelocity()
		e.Physics.AngularVelocity = body.AngularVelocity()
	}
}
