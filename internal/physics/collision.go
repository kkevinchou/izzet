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
	stableSupport           bool
}

const (
	contactPlaneSlop = 1e-6
	satTieTolerance  = 1e-7
)

type cubeSATAxisType int

const (
	cubeSATFaceA cubeSATAxisType = iota
	cubeSATFaceB
	cubeSATEdge
)

type cubeSATResult struct {
	normal      mgl64.Vec3
	penetration float64
	axisType    cubeSATAxisType
	axisA       int
	axisB       int
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
	sat, ok := cubeCubeSAT(a, b)
	if !ok {
		return nil
	}

	var points []mgl64.Vec3
	stableSupport := false
	switch sat.axisType {
	case cubeSATFaceA:
		points = cubeFaceContacts(a, b, sat.normal)
		stableSupport = len(points) >= 3
	case cubeSATFaceB:
		points = cubeFaceContacts(b, a, sat.normal.Mul(-1))
		stableSupport = len(points) >= 3
	case cubeSATEdge:
		a0, a1 := cubeSupportEdge(a, sat.axisA, sat.normal)
		b0, b1 := cubeSupportEdge(b, sat.axisB, sat.normal.Mul(-1))
		pointA, pointB := closestPointsOnSegments(a0, a1, b0, b1)
		points = []mgl64.Vec3{pointA.Add(pointB).Mul(0.5)}
	}

	if len(points) == 0 {
		points = []mgl64.Vec3{cubeCubeContactPoint(a, b, sat.normal)}
	}

	contacts := make([]contact, 0, len(points))
	positionCorrectionScale := 1 / float64(len(points))
	for _, point := range points {
		contact := newContact(a, b, sat.normal, point, sat.penetration)
		contact.positionCorrectionScale = positionCorrectionScale
		contact.stableSupport = stableSupport
		contacts = append(contacts, contact)
	}
	return contacts
}

func cubeCubeSAT(a, b *Body) (cubeSATResult, bool) {
	axesA := cubeAxes(a)
	axesB := cubeAxes(b)
	centerDelta := b.position.Sub(a.position)

	result := cubeSATResult{penetration: math.Inf(1)}
	recordAxis := func(axis mgl64.Vec3, axisType cubeSATAxisType, axisA, axisB int) bool {
		if axis.LenSqr() <= epsilon {
			return true
		}
		axis = axis.Normalize()
		overlap, ok := cubeProjectionOverlap(a, b, axis)
		if !ok {
			return false
		}
		if overlap < result.penetration-satTieTolerance {
			result = cubeSATResult{
				normal:      axis,
				penetration: overlap,
				axisType:    axisType,
				axisA:       axisA,
				axisB:       axisB,
			}
		}
		return true
	}

	for i, axis := range axesA {
		if !recordAxis(axis, cubeSATFaceA, i, -1) {
			return cubeSATResult{}, false
		}
	}
	for i, axis := range axesB {
		if !recordAxis(axis, cubeSATFaceB, -1, i) {
			return cubeSATResult{}, false
		}
	}

	for i, axisA := range axesA {
		for j, axisB := range axesB {
			cross := axisA.Cross(axisB)
			if cross.LenSqr() <= epsilon {
				continue
			}
			if !recordAxis(cross, cubeSATEdge, i, j) {
				return cubeSATResult{}, false
			}
		}
	}

	if result.normal.LenSqr() <= epsilon {
		return cubeSATResult{}, false
	}
	if result.normal.Dot(centerDelta) < 0 {
		result.normal = result.normal.Mul(-1)
	}

	return result, true
}

func cubeProjectionOverlap(a, b *Body, axis mgl64.Vec3) (float64, bool) {
	axis = safeNormalize(axis, mgl64.Vec3{0, 1, 0})
	minA, maxA := projectCube(a, axis)
	minB, maxB := projectCube(b, axis)
	overlap := math.Min(maxA, maxB) - math.Max(minA, minB)
	if overlap < 0 {
		return 0, overlap >= -satTieTolerance
	}
	return overlap, true
}

func cubeFaceContacts(reference, incident *Body, referenceToIncidentNormal mgl64.Vec3) []mgl64.Vec3 {
	faceNormal, faceCenter, sideAxes, sideExtents := cubeFaceFrame(reference, referenceToIncidentNormal)
	polygon := cubeFaceVertices(incident, referenceToIncidentNormal.Mul(-1))
	for i, sideAxis := range sideAxes {
		extent := sideExtents[i]
		polygon = clipPolygonAgainstPlane(polygon, reference.position.Sub(sideAxis.Mul(extent)), sideAxis)
		if len(polygon) == 0 {
			return nil
		}
		polygon = clipPolygonAgainstPlane(polygon, reference.position.Add(sideAxis.Mul(extent)), sideAxis.Mul(-1))
		if len(polygon) == 0 {
			return nil
		}
	}

	points := make([]mgl64.Vec3, 0, len(polygon))
	seen := map[[3]int]bool{}
	for _, vertex := range polygon {
		separation := vertex.Sub(faceCenter).Dot(faceNormal)
		if separation > positionSlop {
			continue
		}
		point := vertex.Sub(faceNormal.Mul(separation * 0.5))
		key := quantizedPointKey(point)
		if seen[key] {
			continue
		}
		seen[key] = true
		points = append(points, point)
	}

	return points
}

func cubeFaceFrame(body *Body, worldNormal mgl64.Vec3) (mgl64.Vec3, mgl64.Vec3, [2]mgl64.Vec3, [2]float64) {
	axes := cubeAxes(body)
	localNormal := body.rotation.Conjugate().Rotate(worldNormal)
	faceAxis := dominantAxis(localNormal)
	sign := 1.0
	if localNormal[faceAxis] < 0 {
		sign = -1
	}

	sideAxis0, sideAxis1 := cubeFaceSideAxisIndices(faceAxis)
	normal := axes[faceAxis].Mul(sign)
	center := body.position.Add(normal.Mul(body.halfExtents[faceAxis]))
	sideAxes := [2]mgl64.Vec3{axes[sideAxis0], axes[sideAxis1]}
	sideExtents := [2]float64{body.halfExtents[sideAxis0], body.halfExtents[sideAxis1]}
	return normal, center, sideAxes, sideExtents
}

func cubeFaceVertices(body *Body, worldNormal mgl64.Vec3) []mgl64.Vec3 {
	axes := cubeAxes(body)
	localNormal := body.rotation.Conjugate().Rotate(worldNormal)
	faceAxis := dominantAxis(localNormal)
	sign := 1.0
	if localNormal[faceAxis] < 0 {
		sign = -1
	}

	sideAxis0, sideAxis1 := cubeFaceSideAxisIndices(faceAxis)
	center := body.position.Add(axes[faceAxis].Mul(body.halfExtents[faceAxis] * sign))
	corners := [][2]float64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
	vertices := make([]mgl64.Vec3, 0, 4)
	for _, corner := range corners {
		vertex := center.
			Add(axes[sideAxis0].Mul(body.halfExtents[sideAxis0] * corner[0])).
			Add(axes[sideAxis1].Mul(body.halfExtents[sideAxis1] * corner[1]))
		vertices = append(vertices, vertex)
	}
	return vertices
}

func cubeFaceSideAxisIndices(faceAxis int) (int, int) {
	switch faceAxis {
	case 0:
		return 1, 2
	case 1:
		return 0, 2
	default:
		return 0, 1
	}
}

func clipPolygonAgainstPlane(points []mgl64.Vec3, planePoint, inwardNormal mgl64.Vec3) []mgl64.Vec3 {
	if len(points) == 0 {
		return nil
	}

	clipped := make([]mgl64.Vec3, 0, len(points)+1)
	previous := points[len(points)-1]
	previousDistance := previous.Sub(planePoint).Dot(inwardNormal)
	previousInside := previousDistance >= -contactPlaneSlop
	for _, current := range points {
		currentDistance := current.Sub(planePoint).Dot(inwardNormal)
		currentInside := currentDistance >= -contactPlaneSlop
		if currentInside != previousInside {
			denominator := previousDistance - currentDistance
			if math.Abs(denominator) > epsilon {
				t := previousDistance / denominator
				clipped = append(clipped, previous.Add(current.Sub(previous).Mul(t)))
			}
		}
		if currentInside {
			clipped = append(clipped, current)
		}

		previous = current
		previousDistance = currentDistance
		previousInside = currentInside
	}
	return clipped
}

func cubeSupportEdge(body *Body, edgeAxis int, direction mgl64.Vec3) (mgl64.Vec3, mgl64.Vec3) {
	axes := cubeAxes(body)
	localDirection := body.rotation.Conjugate().Rotate(direction)
	center := body.position
	for axis := 0; axis < 3; axis++ {
		if axis == edgeAxis {
			continue
		}
		sign := 1.0
		if localDirection[axis] < 0 {
			sign = -1
		}
		center = center.Add(axes[axis].Mul(body.halfExtents[axis] * sign))
	}

	halfEdge := axes[edgeAxis].Mul(body.halfExtents[edgeAxis])
	return center.Sub(halfEdge), center.Add(halfEdge)
}

func closestPointsOnSegments(p1, q1, p2, q2 mgl64.Vec3) (mgl64.Vec3, mgl64.Vec3) {
	d1 := q1.Sub(p1)
	d2 := q2.Sub(p2)
	r := p1.Sub(p2)
	a := d1.Dot(d1)
	e := d2.Dot(d2)
	f := d2.Dot(r)

	if a <= epsilon && e <= epsilon {
		return p1, p2
	}

	var s, t float64
	if a <= epsilon {
		s = 0
		t = clamp(f/e, 0, 1)
	} else {
		c := d1.Dot(r)
		if e <= epsilon {
			t = 0
			s = clamp(-c/a, 0, 1)
		} else {
			b := d1.Dot(d2)
			denominator := a*e - b*b
			if math.Abs(denominator) > epsilon {
				s = clamp((b*f-c*e)/denominator, 0, 1)
			} else {
				s = 0
			}

			tNom := b*s + f
			if tNom < 0 {
				t = 0
				s = clamp(-c/a, 0, 1)
			} else if tNom > e {
				t = 1
				s = clamp((b-c)/a, 0, 1)
			} else {
				t = tNom / e
			}
		}
	}

	return p1.Add(d1.Mul(s)), p2.Add(d2.Mul(t))
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
