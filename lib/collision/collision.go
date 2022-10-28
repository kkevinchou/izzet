package collision

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/utils"
	"github.com/kkevinchou/izzet/lib/collision/checks"
	"github.com/kkevinchou/izzet/lib/collision/collider"
)

type ContactsBySeparatingDistance []*Contact

func (c ContactsBySeparatingDistance) Len() int {
	return len(c)
}
func (c ContactsBySeparatingDistance) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c ContactsBySeparatingDistance) Less(i, j int) bool {
	return c[i].SeparatingDistance < c[j].SeparatingDistance
}

type ContactType string

var ContactTypeCapsuleTriMesh ContactType = "TRIMESH"
var ContactTypeCapsuleCapsule ContactType = "CAPSULE"

type Contact struct {
	EntityID       *int
	SourceEntityID *int
	Type           ContactType

	TriIndex           *int
	Point              mgl64.Vec3
	Normal             mgl64.Vec3
	SeparatingVector   mgl64.Vec3
	SeparatingDistance float64
}

func (c *Contact) String() string {
	var result = fmt.Sprintf("{ EntityID: %d, TriIndex: %d ", *c.EntityID, *c.TriIndex)
	result += fmt.Sprintf("[ SV: %s, N: %s, D: %.3f D2: %.3f]", utils.PPrintVec(c.SeparatingVector), utils.PPrintVec(c.Normal), c.SeparatingDistance, c.SeparatingVector.Len())
	result += " }"
	return result
}

func CheckCollisionCapsuleTriMesh(capsule collider.Capsule, triangulatedMesh collider.TriMesh) []*Contact {
	var contacts []*Contact
	for i, tri := range triangulatedMesh.Triangles {
		if triContact := CheckCollisionCapsuleTriangle(capsule, tri); triContact != nil {
			index := i
			triContact.TriIndex = &index
			contacts = append(contacts, triContact)
		}
	}

	return contacts
}

// func CheckCollisionLineTriangle(line collider.Line, triangle collider.Triangle) *Contact {
// 	dir1 := line.P1.Sub(line.P2)
// 	dir2 := line.P2.Sub(line.P1)

// 	return nil
// }

func CheckCollisionCapsuleTriangle(capsule collider.Capsule, triangle collider.Triangle) *Contact {
	closestPoints, closestPointsDistance := checks.ClosestPointsLineVSTriangle(
		collider.Line{P1: capsule.Top, P2: capsule.Bottom},
		triangle,
	)
	// closestPointCapsule := closestPoints[0]
	// closestPointTriangle := closestPoints[1]

	if closestPointsDistance < capsule.Radius {
		separatingDistance := capsule.Radius - closestPointsDistance
		separatingVec := closestPoints[0].Sub(closestPoints[1]).Normalize().Mul(separatingDistance)
		if separatingVec.Dot(triangle.Normal) < 0 {
			// TODO(kevin): not sure if this is right, might want to revisit
			// hacky handling of separating vector pushing the capsule opposite to the triangle normal
			separatingVec = separatingVec.Add(triangle.Normal.Mul(capsule.Radius * 2))
			separatingDistance = separatingVec.Len()
		}
		return &Contact{
			Point:              closestPoints[1],
			Normal:             triangle.Normal,
			SeparatingVector:   separatingVec,
			SeparatingDistance: separatingDistance,
			Type:               ContactTypeCapsuleTriMesh,
		}
	}

	return nil
}

// for now assumes vertical capsules only
func CheckCollisionCapsuleCapsule(capsule1 collider.Capsule, capsule2 collider.Capsule) *Contact {
	closestPoints, closestPointsDistance := checks.ClosestPointsLineVSLine(
		collider.Line{P1: capsule1.Top, P2: capsule1.Bottom},
		collider.Line{P1: capsule2.Top, P2: capsule2.Bottom},
	)

	separatingDistance := (capsule1.Radius + capsule2.Radius) - closestPointsDistance
	if separatingDistance > 0 {
		capsule2To1 := closestPoints[0].Sub(closestPoints[1]).Normalize()
		separatingVec := capsule2To1.Mul(separatingDistance)
		return &Contact{
			Point:              capsule2To1.Mul(closestPointsDistance),
			SeparatingVector:   separatingVec,
			SeparatingDistance: separatingDistance,
			Type:               ContactTypeCapsuleCapsule,
		}
	}

	return nil
}

func CheckOverlapAABBAABB(aabb1 *collider.BoundingBox, aabb2 *collider.BoundingBox) bool {
	if aabb1.MaxVertex.X() < aabb2.MinVertex.X() || aabb1.MinVertex.X() > aabb2.MaxVertex.X() {
		return false
	}

	if aabb1.MaxVertex.Y() < aabb2.MinVertex.Y() || aabb1.MinVertex.Y() > aabb2.MaxVertex.Y() {
		return false
	}

	if aabb1.MaxVertex.Z() < aabb2.MinVertex.Z() || aabb1.MinVertex.Z() > aabb2.MaxVertex.Z() {
		return false
	}

	return true
}

// func CheckCollisionSpherePoint(sphere collider.Sphere, point mgl64.Vec3) *ContactManifold {
// 	lenSq := sphere.Center.Sub(mgl64.Vec3(point)).LenSqr()
// 	if lenSq < sphere.RadiusSquared {
// 		return &ContactManifold{
// 			Contacts: []Contact{
// 				{
// 					Point: mgl64.Vec3{point[0], point[1], point[2]},
// 					// Normal: sphere.Center.Sub(mgl64.Vec3(point)),
// 				},
// 			},
// 		}
// 	}

// 	return nil
// }
