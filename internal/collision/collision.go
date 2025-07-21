package collision

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/collision/checks"
	"github.com/kkevinchou/kitolib/collision/collider"
)

type Contact struct {
	Type ContactType

	SeparatingVector   mgl64.Vec3
	SeparatingDistance float64
}

type ContactType string

var ContactTypeCapsuleTriMesh ContactType = "TRIMESH"
var ContactTypeCapsuleCapsule ContactType = "CAPSULE"

// for now assumes vertical capsules only
func CheckCollisionCapsuleCapsule(capsule1 collider.Capsule, capsule2 collider.Capsule) (Contact, bool) {
	closestPoints, closestPointsDistance := checks.ClosestPointsLineVSLine(
		collider.Line{P1: capsule1.Top, P2: capsule1.Bottom},
		collider.Line{P1: capsule2.Top, P2: capsule2.Bottom},
	)

	separatingDistance := (capsule1.Radius + capsule2.Radius) - closestPointsDistance
	if separatingDistance > 0 {
		capsule2To1 := closestPoints[0].Sub(closestPoints[1])

		// if the two capsules are directly atop one another, push the capsule up
		if capsule2To1.LenSqr() == 0 {
			separatingDistance = capsule2.Top.Sub(capsule2.Bottom).Len() + 2*capsule2.Radius
			capsule2To1 = mgl64.Vec3{0, 1, 0}
		}

		capsule2To1Dir := capsule2To1.Normalize()
		separatingVec := capsule2To1Dir.Mul(separatingDistance)
		return Contact{
			// Point:              capsule2To1Dir.Mul(closestPointsDistance),
			SeparatingVector:   separatingVec,
			SeparatingDistance: separatingDistance,
			Type:               ContactTypeCapsuleCapsule,
		}, true
	}

	return Contact{}, false
}

func CheckCollisionCapsuleTriMesh(capsule collider.Capsule, triangulatedMesh collider.TriMesh) []Contact {
	var contacts []Contact
	for _, tri := range triangulatedMesh.Triangles {
		if triContact, collision := CheckCollisionCapsuleTriangle(capsule, tri); collision {
			// index := i
			// triContact.TriIndex = &index
			contacts = append(contacts, triContact)
		}
	}

	return contacts
}

func CheckCollisionCapsuleTriangle(capsule collider.Capsule, triangle collider.Triangle) (Contact, bool) {
	closestPoints, closestPointsDistance := checks.ClosestPointsLineVSTriangle(
		collider.Line{P1: capsule.Top, P2: capsule.Bottom},
		triangle,
	)

	if closestPointsDistance == 0 {
		// separating vector of length 0
		return Contact{}, false
	}

	if closestPointsDistance < capsule.Radius {
		separatingDistance := capsule.Radius - closestPointsDistance
		separatingVec := closestPoints[0].Sub(closestPoints[1]).Normalize().Mul(separatingDistance)
		if separatingVec.Dot(triangle.Normal) < 0 {
			// TODO(kevin): not sure if this is right, might want to revisit
			// hacky handling of separating vector pushing the capsule opposite to the triangle normal
			separatingVec = separatingVec.Add(triangle.Normal.Mul(capsule.Radius * 2))
			separatingDistance = separatingVec.Len()
		}
		return Contact{
			// Point:              closestPoints[1],
			// Normal:             triangle.Normal,
			SeparatingVector:   separatingVec,
			SeparatingDistance: separatingDistance,
			Type:               ContactTypeCapsuleTriMesh,
		}, true
	}

	return Contact{}, false
}
