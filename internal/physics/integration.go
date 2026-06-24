package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

func (w *World) integrate(dt float64) {
	for _, id := range w.bodyOrder {
		body, ok := w.bodies[id]
		if !ok || body.Static() {
			continue
		}

		linearAcceleration := w.gravity.Add(body.force.Mul(body.inverseMass))
		body.linearVelocity = body.linearVelocity.Add(linearAcceleration.Mul(dt))

		angularAcceleration := body.applyInverseInertia(body.torque)
		body.angularVelocity = body.angularVelocity.Add(angularAcceleration.Mul(dt))

		body.linearVelocity = body.linearVelocity.Mul(dampingFactor(body.linearDamping, dt))
		body.angularVelocity = body.angularVelocity.Mul(dampingFactor(body.angularDamping, dt))

		body.position = body.position.Add(body.linearVelocity.Mul(dt))
		body.rotation = integrateRotation(body.rotation, body.angularVelocity, dt)
		body.ClearForces()
	}
}

func dampingFactor(damping, dt float64) float64 {
	if damping <= 0 {
		return 1
	}
	if damping >= 1 {
		return 0
	}
	return math.Pow(1-damping, dt)
}

func integrateRotation(rotation mgl64.Quat, angularVelocity mgl64.Vec3, dt float64) mgl64.Quat {
	angle := angularVelocity.Len() * dt
	if angle <= epsilon {
		return rotation
	}

	delta := mgl64.QuatRotate(angle, angularVelocity.Normalize())
	return delta.Mul(rotation).Normalize()
}
