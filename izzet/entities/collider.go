package entities

import (
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/collision/collider"
)

type ColliderComponent struct {
	// entities with the same collider group do not collide with each other
	ColliderGroup types.ColliderGroupFlag
	CollisionMask types.ColliderGroupFlag

	// Skip separation tells the collision system to skip the step of separating colliding entities
	// for the entity that owns this component
	SkipSeparation bool

	// Contacts marks which entities it collided with in the current frame
	Contacts []collision.Contact

	CapsuleCollider           *collider.Capsule
	TriMeshCollider           *collider.TriMesh `json:"-"`
	SimplifiedTriMeshCollider *collider.TriMesh `json:"-"`
	BoundingBoxCollider       *collider.BoundingBox

	// stores the transformed collider (e.g. if the entity moves)
	TransformedCapsuleCollider     *collider.Capsule     `json:"-"`
	TransformedTriMeshCollider     *collider.TriMesh     `json:"-"`
	TransformedBoundingBoxCollider *collider.BoundingBox `json:"-"`
}
