package system

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
)

type PhysicsSystem struct {
	app App
}

func NewPhysicsSystem(app App) *PhysicsSystem {
	return &PhysicsSystem{app: app}
}

func (s *PhysicsSystem) Name() string {
	return "PhysicsSystem"
}

func (s *PhysicsSystem) Update(delta time.Duration, world GameWorld) {
	if !s.app.IsServer() {
		return
	}

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
