package world

import (
	"math"

	phys "github.com/kkevinchou/izzet/internal/physics"
	"github.com/kkevinchou/izzet/izzet/entity"
)

func (g *GameWorld) addPhysicsBody(e *entity.Entity) error {
	if e.Physics == nil {
		return nil
	}

	if e.Physics.BodyID != 0 {
		if _, ok := g.PhysicsWorld().Body(e.Physics.BodyID); ok {
			return nil
		}
		e.Physics.BodyID = 0
	}

	options := physicsBodyOptions(e)
	var bodyID phys.BodyID
	var err error

	switch physicsShape(e.Physics) {
	case entity.PhysicsShapeSphere:
		bodyID, err = g.PhysicsWorld().CreateSphereWithOptions(phys.SphereOptions{
			BodyOptions: options,
			Radius:      physicsSphereRadius(e),
		})
	default:
		bodyID, err = g.PhysicsWorld().CreateCubeWithOptions(phys.CubeOptions{
			BodyOptions: options,
			Size:        e.Scale(),
		})
	}
	if err != nil {
		return err
	}

	e.Physics.BodyID = bodyID
	return nil
}

func physicsBodyOptions(e *entity.Entity) phys.BodyOptions {
	options := phys.DefaultBodyOptions(e.Physics.Mass)
	options.Position = e.Position()
	options.Rotation = e.Rotation()
	options.LinearVelocity = e.Physics.Velocity
	options.AngularVelocity = e.Physics.AngularVelocity
	options.Static = e.Static

	if e.Physics.Restitution != 0 {
		options.Restitution = e.Physics.Restitution
	}
	if e.Physics.Friction != 0 {
		options.Friction = e.Physics.Friction
	}
	if e.Physics.LinearDamping != 0 {
		options.LinearDamping = e.Physics.LinearDamping
	}
	if e.Physics.AngularDamping != 0 {
		options.AngularDamping = e.Physics.AngularDamping
	}

	return options
}

func physicsShape(component *entity.PhysicsComponent) entity.PhysicsShape {
	if component.Shape == "" {
		return entity.PhysicsShapeCube
	}
	return component.Shape
}

func physicsSphereRadius(e *entity.Entity) float64 {
	if e.Physics.Radius > 0 {
		return e.Physics.Radius
	}

	scale := e.Scale()
	radius := math.Max(math.Abs(scale.X()), math.Max(math.Abs(scale.Y()), math.Abs(scale.Z()))) * 0.5
	return radius
}
