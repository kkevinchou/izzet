package netsync

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/libutils"
)

const (
	fullDecayThreshold float64 = 0.05
	minYPosition       float64 = -1000
)

func PhysicsStep(delta time.Duration, entity entities.Entity) {
	componentContainer := entity.GetComponentContainer()
	physicsComponent := componentContainer.PhysicsComponent
	transformComponent := componentContainer.TransformComponent

	if physicsComponent.Static {
		return
	}

	// calculate impulses and their decay, this is meant for controller
	// actions that can "overwite" impulses
	var totalImpulse mgl64.Vec3
	for name, impulse := range physicsComponent.Impulses {
		decayRatio := 1.0 - (impulse.ElapsedTime.Seconds() * impulse.DecayRate)
		if decayRatio < 0 {
			decayRatio = 0
		}

		if decayRatio < fullDecayThreshold {
			delete(physicsComponent.Impulses, name)
			continue
		} else {
			realImpulse := impulse.Vector.Mul(decayRatio)
			totalImpulse = totalImpulse.Add(realImpulse)
		}

		// update the impulse
		impulse.ElapsedTime = impulse.ElapsedTime + delta
		physicsComponent.Impulses[name] = impulse
	}

	// calculate velocity adjusted by acceleration
	var totalAcceleration mgl64.Vec3
	if !physicsComponent.IgnoreGravity {
		totalAcceleration = totalAcceleration.Add(settings.AccelerationDueToGravity)
	}
	physicsComponent.Velocity = physicsComponent.Velocity.Add(totalAcceleration.Mul(delta.Seconds()))

	velocity := physicsComponent.Velocity.Add(totalImpulse)
	newPos := transformComponent.Position.Add(velocity.Mul(delta.Seconds()))

	// temporary hack to not fall through the ground
	if newPos[1] < minYPosition {
		newPos[1] = 0
		velocity[1] = 0
		physicsComponent.Velocity[1] = 0
		delete(physicsComponent.Impulses, types.JumpImpulse)
	}

	transformComponent.Position = newPos

	// updating orientation along velocity
	velocityWithoutY := mgl64.Vec3{velocity[0], 0, velocity[2]}
	if !libutils.Vec3IsZero(velocityWithoutY) {
		// Note, this will bug out if we look directly up or directly down. This
		// is due to issues looking at objects that are along our "up" vector.
		// I believe this is due to us losing sense of what a "right" vector is.
		// This code will likely change when we do animation blending in the animator
		transformComponent.Orientation = mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, velocityWithoutY.Normalize())
	}
}
