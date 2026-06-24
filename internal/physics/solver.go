package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

const (
	positionSlop       = 0.001
	positionCorrection = 0.8
	restingVelocity    = 1.0
	linearSleepSpeed   = 0.03
	angularSleepSpeed  = 0.03
)

func resolveContactVelocity(c *contact) {
	relativeVelocity := velocityAtPoint(c.b, c.point).Sub(velocityAtPoint(c.a, c.point))
	velocityAlongNormal := relativeVelocity.Dot(c.normal)
	if velocityAlongNormal > 0 {
		return
	}

	restitution := math.Min(c.a.restitution, c.b.restitution)
	if velocityAlongNormal > -restingVelocity {
		restitution = 0
	}

	denominator := impulseDenominator(c.a, c.b, c.point, c.normal)
	if denominator <= epsilon {
		return
	}

	j := -(1 + restitution) * velocityAlongNormal / denominator
	normalImpulse := c.normal.Mul(j)
	c.a.applyImpulse(normalImpulse.Mul(-1), c.point)
	c.b.applyImpulse(normalImpulse, c.point)

	relativeVelocity = velocityAtPoint(c.b, c.point).Sub(velocityAtPoint(c.a, c.point))
	tangent := relativeVelocity.Sub(c.normal.Mul(relativeVelocity.Dot(c.normal)))
	if tangent.LenSqr() <= epsilon {
		return
	}

	tangent = tangent.Normalize()
	tangentDenominator := impulseDenominator(c.a, c.b, c.point, tangent)
	if tangentDenominator <= epsilon {
		return
	}

	jt := -relativeVelocity.Dot(tangent) / tangentDenominator
	friction := math.Sqrt(c.a.friction * c.b.friction)
	maxFriction := math.Abs(j) * friction
	jt = clamp(jt, -maxFriction, maxFriction)

	frictionImpulse := tangent.Mul(jt)
	c.a.applyImpulse(frictionImpulse.Mul(-1), c.point)
	c.b.applyImpulse(frictionImpulse, c.point)
}

func correctContactPosition(c *contact) {
	inverseMassSum := c.a.inverseMass + c.b.inverseMass
	if inverseMassSum <= epsilon {
		return
	}

	positionCorrectionScale := c.positionCorrectionScale
	if positionCorrectionScale == 0 {
		positionCorrectionScale = 1
	}
	magnitude := math.Max(c.penetration-positionSlop, 0) / inverseMassSum * positionCorrection * positionCorrectionScale
	correction := c.normal.Mul(magnitude)

	if !c.a.Static() {
		c.a.position = c.a.position.Sub(correction.Mul(c.a.inverseMass))
	}
	if !c.b.Static() {
		c.b.position = c.b.position.Add(correction.Mul(c.b.inverseMass))
	}
}

func velocityAtPoint(body *Body, worldPoint mgl64.Vec3) mgl64.Vec3 {
	r := worldPoint.Sub(body.position)
	return body.linearVelocity.Add(body.angularVelocity.Cross(r))
}

func impulseDenominator(a, b *Body, point, normal mgl64.Vec3) float64 {
	ra := point.Sub(a.position)
	rb := point.Sub(b.position)

	angularA := a.applyInverseInertia(ra.Cross(normal)).Cross(ra)
	angularB := b.applyInverseInertia(rb.Cross(normal)).Cross(rb)

	return a.inverseMass + b.inverseMass + normal.Dot(angularA.Add(angularB))
}

func stabilizeRestingContacts(contacts []contact) {
	for i := range contacts {
		if contacts[i].positionCorrectionScale > 1.0/3.0 {
			continue
		}

		a := contacts[i].a
		b := contacts[i].b

		if a.Static() && !b.Static() {
			stabilizeRestingBody(b)
		} else if b.Static() && !a.Static() {
			stabilizeRestingBody(a)
		}
	}
}

func stabilizeRestingBody(body *Body) {
	if body.linearVelocity.LenSqr() < linearSleepSpeed*linearSleepSpeed {
		body.linearVelocity = mgl64.Vec3{}
	}
	if body.angularVelocity.LenSqr() < angularSleepSpeed*angularSleepSpeed {
		body.angularVelocity = mgl64.Vec3{}
	}
}
