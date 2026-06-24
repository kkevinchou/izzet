package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

type contact struct {
	a                       *Body
	b                       *Body
	normal                  mgl64.Vec3
	point                   mgl64.Vec3
	penetration             float64
	positionCorrectionScale float64
}

func (w *World) detectContacts() []contact {
	bodies := w.Bodies()
	contacts := make([]contact, 0)

	for i := 0; i < len(bodies); i++ {
		for j := i + 1; j < len(bodies); j++ {
			a := bodies[i]
			b := bodies[j]
			if a.Static() && b.Static() {
				continue
			}
			if !boundingSpheresOverlap(a, b) {
				continue
			}
			contacts = append(contacts, detectContactsBetween(a, b)...)
		}
	}

	return contacts
}

func boundingSpheresOverlap(a, b *Body) bool {
	radius := a.BoundingRadius() + b.BoundingRadius()
	return b.position.Sub(a.position).LenSqr() <= radius*radius
}

func detectContactsBetween(a, b *Body) []contact {
	switch {
	case a.shape == ShapeSphere && b.shape == ShapeSphere:
		normal, point, penetration, ok := sphereSphereContact(a, b)
		if !ok {
			return nil
		}
		return []contact{newContact(a, b, normal, point, penetration)}
	case a.shape == ShapeSphere && b.shape == ShapeCube:
		normalCubeToSphere, point, penetration, ok := sphereCubeContact(a, b)
		if !ok {
			return nil
		}
		return []contact{newContact(a, b, normalCubeToSphere.Mul(-1), point, penetration)}
	case a.shape == ShapeCube && b.shape == ShapeSphere:
		normalCubeToSphere, point, penetration, ok := sphereCubeContact(b, a)
		if !ok {
			return nil
		}
		return []contact{newContact(a, b, normalCubeToSphere, point, penetration)}
	case a.shape == ShapeCube && b.shape == ShapeCube:
		return cubeCubeContacts(a, b)
	default:
		return nil
	}
}

func newContact(a, b *Body, normal, point mgl64.Vec3, penetration float64) contact {
	return contact{
		a:                       a,
		b:                       b,
		normal:                  normal,
		point:                   point,
		penetration:             penetration,
		positionCorrectionScale: 1,
	}
}

func sphereSphereContact(a, b *Body) (mgl64.Vec3, mgl64.Vec3, float64, bool) {
	delta := b.position.Sub(a.position)
	distanceSquared := delta.LenSqr()
	radius := a.radius + b.radius
	if distanceSquared > radius*radius {
		return mgl64.Vec3{}, mgl64.Vec3{}, 0, false
	}

	distance := math.Sqrt(distanceSquared)
	normal := safeNormalize(delta, mgl64.Vec3{0, 1, 0})
	penetration := radius - distance
	pointA := a.position.Add(normal.Mul(a.radius))
	pointB := b.position.Sub(normal.Mul(b.radius))
	return normal, pointA.Add(pointB).Mul(0.5), penetration, true
}

// sphereCubeContact returns the contact normal from the cube toward the sphere.
func sphereCubeContact(sphere, cube *Body) (mgl64.Vec3, mgl64.Vec3, float64, bool) {
	invRotation := cube.rotation.Conjugate()
	localSphereCenter := invRotation.Rotate(sphere.position.Sub(cube.position))
	half := cube.halfExtents

	clamped := mgl64.Vec3{
		clamp(localSphereCenter.X(), -half.X(), half.X()),
		clamp(localSphereCenter.Y(), -half.Y(), half.Y()),
		clamp(localSphereCenter.Z(), -half.Z(), half.Z()),
	}

	localDelta := localSphereCenter.Sub(clamped)
	distanceSquared := localDelta.LenSqr()
	if distanceSquared > sphere.radius*sphere.radius {
		return mgl64.Vec3{}, mgl64.Vec3{}, 0, false
	}

	if distanceSquared > epsilon {
		distance := math.Sqrt(distanceSquared)
		normal := cube.rotation.Rotate(localDelta.Mul(1 / distance))
		closestWorld := cube.position.Add(cube.rotation.Rotate(clamped))
		spherePoint := sphere.position.Sub(normal.Mul(sphere.radius))
		return normal, closestWorld.Add(spherePoint).Mul(0.5), sphere.radius - distance, true
	}

	normalLocal, distanceToFace := closestCubeFaceNormal(localSphereCenter, half)
	normal := cube.rotation.Rotate(normalLocal)
	closestLocal := localSphereCenter
	axis := dominantAxis(normalLocal)
	closestLocal[axis] = half[axis] * math.Copysign(1, normalLocal[axis])
	closestWorld := cube.position.Add(cube.rotation.Rotate(closestLocal))
	spherePoint := sphere.position.Sub(normal.Mul(sphere.radius))
	return normal, closestWorld.Add(spherePoint).Mul(0.5), sphere.radius + distanceToFace, true
}

func closestCubeFaceNormal(localPoint, halfExtents mgl64.Vec3) (mgl64.Vec3, float64) {
	distances := halfExtents.Sub(componentAbs(localPoint))
	axis := 0
	distance := distances.X()
	if distances.Y() < distance {
		axis = 1
		distance = distances.Y()
	}
	if distances.Z() < distance {
		axis = 2
		distance = distances.Z()
	}

	normal := mgl64.Vec3{}
	sign := 1.0
	if localPoint[axis] < 0 {
		sign = -1
	}
	normal[axis] = sign
	return normal, distance
}

func dominantAxis(v mgl64.Vec3) int {
	abs := componentAbs(v)
	if abs.X() >= abs.Y() && abs.X() >= abs.Z() {
		return 0
	}
	if abs.Y() >= abs.Z() {
		return 1
	}
	return 2
}

func cubeCubeContacts(a, b *Body) []contact {
	normal, penetration, ok := cubeCubeSAT(a, b)
	if !ok {
		return nil
	}

	points := cubeCubeContactPoints(a, b, normal, penetration)
	if len(points) == 0 {
		points = []mgl64.Vec3{cubeCubeContactPoint(a, b, normal)}
	}

	contacts := make([]contact, 0, len(points))
	positionCorrectionScale := 1 / float64(len(points))
	for _, point := range points {
		contact := newContact(a, b, normal, point, penetration)
		contact.positionCorrectionScale = positionCorrectionScale
		contacts = append(contacts, contact)
	}
	return contacts
}

func cubeCubeSAT(a, b *Body) (mgl64.Vec3, float64, bool) {
	axesA := cubeAxes(a)
	axesB := cubeAxes(b)

	centerDelta := b.position.Sub(a.position)
	minOverlap := math.Inf(1)
	var bestAxis mgl64.Vec3

	for _, axis := range append(axesA[:], axesB[:]...) {
		overlap, ok := cubeProjectionOverlap(a, b, axis)
		if !ok {
			return mgl64.Vec3{}, 0, false
		}
		if overlap < minOverlap {
			minOverlap = overlap
			bestAxis = axis
		}
	}

	edgeAxisBias := math.Max(0.02, minOverlap*0.05)
	for _, axisA := range axesA {
		for _, axisB := range axesB {
			cross := axisA.Cross(axisB)
			if cross.LenSqr() <= epsilon {
				continue
			}

			axis := cross.Normalize()
			overlap, ok := cubeProjectionOverlap(a, b, axis)
			if !ok {
				return mgl64.Vec3{}, 0, false
			}
			if overlap < minOverlap-edgeAxisBias {
				minOverlap = overlap
				bestAxis = axis
			}
		}
	}

	if bestAxis.Dot(centerDelta) < 0 {
		bestAxis = bestAxis.Mul(-1)
	}

	return bestAxis, minOverlap, true
}

func cubeProjectionOverlap(a, b *Body, axis mgl64.Vec3) (float64, bool) {
	axis = safeNormalize(axis, mgl64.Vec3{0, 1, 0})
	minA, maxA := projectCube(a, axis)
	minB, maxB := projectCube(b, axis)
	overlap := math.Min(maxA, maxB) - math.Max(minA, minB)
	return overlap, overlap >= 0
}

func cubeCubeContactPoints(a, b *Body, normal mgl64.Vec3, penetration float64) []mgl64.Vec3 {
	reference := a
	incident := b
	referenceNormal := normal
	if b.Static() && !a.Static() {
		reference = b
		incident = a
		referenceNormal = normal.Mul(-1)
	}

	incidentVertices := cubeVertices(incident)
	minProjection := math.Inf(1)
	for _, vertex := range incidentVertices {
		projection := vertex.Dot(referenceNormal)
		if projection < minProjection {
			minProjection = projection
		}
	}

	points := make([]mgl64.Vec3, 0, 4)
	vertexSlop := math.Max(0.02, penetration+0.01)
	seen := map[[3]int]bool{}
	for _, vertex := range incidentVertices {
		if vertex.Dot(referenceNormal)-minProjection > vertexSlop {
			continue
		}

		point := cubeFacePoint(reference, referenceNormal, vertex)
		key := quantizedPointKey(point)
		if seen[key] {
			continue
		}
		seen[key] = true
		points = append(points, point)
	}

	return points
}

func cubeCubeContactPoint(a, b *Body, normal mgl64.Vec3) mgl64.Vec3 {
	pointA := cubeFacePoint(a, normal, b.position)
	pointB := cubeFacePoint(b, normal.Mul(-1), a.position)

	if a.Static() && !b.Static() {
		return pointA
	}
	if b.Static() && !a.Static() {
		return pointB
	}

	return pointA.Add(pointB).Mul(0.5)
}

func cubeFacePoint(body *Body, worldNormal, target mgl64.Vec3) mgl64.Vec3 {
	localTarget := body.rotation.Conjugate().Rotate(target.Sub(body.position))
	localNormal := body.rotation.Conjugate().Rotate(worldNormal)
	axis := dominantAxis(localNormal)

	point := mgl64.Vec3{
		clamp(localTarget.X(), -body.halfExtents.X(), body.halfExtents.X()),
		clamp(localTarget.Y(), -body.halfExtents.Y(), body.halfExtents.Y()),
		clamp(localTarget.Z(), -body.halfExtents.Z(), body.halfExtents.Z()),
	}

	sign := 1.0
	if localNormal[axis] < 0 {
		sign = -1
	}
	point[axis] = body.halfExtents[axis] * sign

	return body.position.Add(body.rotation.Rotate(point))
}

func cubeAxes(body *Body) [3]mgl64.Vec3 {
	return [3]mgl64.Vec3{
		body.rotation.Rotate(mgl64.Vec3{1, 0, 0}),
		body.rotation.Rotate(mgl64.Vec3{0, 1, 0}),
		body.rotation.Rotate(mgl64.Vec3{0, 0, 1}),
	}
}

func cubeVertices(body *Body) []mgl64.Vec3 {
	axes := cubeAxes(body)
	vertices := make([]mgl64.Vec3, 0, 8)
	for _, sx := range []float64{-1, 1} {
		for _, sy := range []float64{-1, 1} {
			for _, sz := range []float64{-1, 1} {
				vertex := body.position.
					Add(axes[0].Mul(body.halfExtents.X() * sx)).
					Add(axes[1].Mul(body.halfExtents.Y() * sy)).
					Add(axes[2].Mul(body.halfExtents.Z() * sz))
				vertices = append(vertices, vertex)
			}
		}
	}
	return vertices
}

func quantizedPointKey(point mgl64.Vec3) [3]int {
	const scale = 10000
	return [3]int{
		int(math.Round(point.X() * scale)),
		int(math.Round(point.Y() * scale)),
		int(math.Round(point.Z() * scale)),
	}
}

func projectCube(body *Body, axis mgl64.Vec3) (float64, float64) {
	axes := cubeAxes(body)
	center := body.position.Dot(axis)
	radius := math.Abs(axis.Dot(axes[0]))*body.halfExtents.X() +
		math.Abs(axis.Dot(axes[1]))*body.halfExtents.Y() +
		math.Abs(axis.Dot(axes[2]))*body.halfExtents.Z()
	return center - radius, center + radius
}

func supportPoint(body *Body, direction mgl64.Vec3) mgl64.Vec3 {
	switch body.shape {
	case ShapeSphere:
		return body.position.Add(safeNormalize(direction, mgl64.Vec3{0, 1, 0}).Mul(body.radius))
	case ShapeCube:
		axes := cubeAxes(body)
		point := body.position
		for i, axis := range axes {
			sign := 1.0
			if direction.Dot(axis) < 0 {
				sign = -1
			}
			point = point.Add(axis.Mul(body.halfExtents[i] * sign))
		}
		return point
	default:
		return body.position
	}
}
