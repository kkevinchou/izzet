package entities

import (
	"github.com/go-gl/mathgl/mgl64"
)

type KinematicComponent struct {
	Velocity            mgl64.Vec3
	AccumulatedVelocity mgl64.Vec3
	Grounded            bool
	GravityEnabled      bool
	RotateOnVelocity    bool
}

func (e *Entity) IsKinematic() bool {
	return e.Kinematic != nil
}

func (e *Entity) IsStatic() bool {
	return e.Static
}

func (e *Entity) GravityEnabled() bool {
	return e.Kinematic.GravityEnabled
}

func (e *Entity) TotalKinematicVelocity() mgl64.Vec3 {
	return e.Kinematic.Velocity.Add(e.Kinematic.AccumulatedVelocity)
}

func (e *Entity) AccumulateKinematicVelocity(v mgl64.Vec3) {
	e.Kinematic.AccumulatedVelocity = e.Kinematic.AccumulatedVelocity.Add(v)
}

func (e *Entity) AddPosition(v mgl64.Vec3) {
	SetLocalPosition(e, e.LocalPosition.Add(v))
}

func (e *Entity) SetPosition(v mgl64.Vec3) {
	SetLocalPosition(e, v)
}

func (e *Entity) ClearKinematicVelocity() {
	e.Kinematic.Velocity = mgl64.Vec3{}
	e.Kinematic.AccumulatedVelocity = mgl64.Vec3{}
}

func (e *Entity) SetGrounded(v bool) {
	e.Kinematic.Grounded = v
}
