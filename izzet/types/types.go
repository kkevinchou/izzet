package types

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
)

type MeshHandle struct {
	Namespace string
	ID        string
}

type MaterialHandle struct {
	ID string
}

func (h MaterialHandle) String() string {
	return h.ID
}

type KinematicEntity interface {
	GetID() int
	IsKinematic() bool
	IsStatic() bool
	GravityEnabled() bool
	TotalKinematicVelocity() mgl64.Vec3
	AccumulateKinematicVelocity(v mgl64.Vec3)
	ClearVerticalKinematicVelocity()
	SetGrounded(v bool)
	Position() mgl64.Vec3
	AddPosition(v mgl64.Vec3)
	SetPosition(v mgl64.Vec3)
	BoundingBox() collider.BoundingBox
	SetLocalRotation(q mgl64.Quat)
	GetMovementVector() mgl64.Vec3

	HasCapsuleCollider() bool
	HasTriMeshCollider() bool
	HasSimplifiedTriMeshCollider() bool
	CapsuleCollider() collider.Capsule
	TriMeshCollider() collider.TriMesh
	SimplifiedTriMeshCollider() collider.TriMesh
	GetLocalRotation() mgl64.Quat
}
