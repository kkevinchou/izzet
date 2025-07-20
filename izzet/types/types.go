package types

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/collision/collider"
)

type MeshHandle struct {
	Namespace string
	ID        string
}

type MaterialHandle struct {
	Namespace string
	ID        string
}

func (h MaterialHandle) String() string {
	return h.Namespace + "-" + h.ID
}

type KinematicEntity interface {
	IsKinematic() bool
	IsStatic() bool
	GravityEnabled() bool
	KinematicVelocity() mgl64.Vec3
	AddKinematicVelocity(v mgl64.Vec3)
	Position() mgl64.Vec3
	AddPosition(v mgl64.Vec3)
	SetPosition(v mgl64.Vec3)
	BoundingBox() collider.BoundingBox

	HasCapsuleCollider() bool
	HasTriMeshCollider() bool
	CapsuleCollider() collider.Capsule
	TriMeshCollider() collider.TriMesh
}
