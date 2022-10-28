package physics

import "github.com/go-gl/mathgl/mgl64"

type ColliderType string

const (
	ColliderTypeCapsule ColliderType = "CAPSULE"
	ColliderTypeBox     ColliderType = "BOX"
	ColliderTypeSphere  ColliderType = "SPHERE"
	// RigidBodyTypeCapsule RigidBodyType = "CAPSULE"
)

type Capsule struct {
	Position mgl64.Vec3
}

type Sphere struct {
	Position mgl64.Vec3
	Radius   float64
}

type Box struct {
}
