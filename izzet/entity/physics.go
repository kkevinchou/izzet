package entity

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/physics"
)

type PhysicsShape string

const (
	PhysicsShapeCube   PhysicsShape = "CUBE"
	PhysicsShapeSphere PhysicsShape = "SPHERE"
)

type PhysicsComponent struct {
	BodyID physics.BodyID `json:"-"`

	Shape  PhysicsShape
	Mass   float64
	Radius float64

	Restitution    float64
	Friction       float64
	LinearDamping  float64
	AngularDamping float64

	Velocity        mgl64.Vec3
	AngularVelocity mgl64.Vec3
}
