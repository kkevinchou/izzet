package components

import (
	"github.com/kkevinchou/kitolib/collision/collider"
)

type ColliderComponent struct {
	// Skip separation tells the collision system to skip the step of separating colliding entities
	// for the entity that owns this component
	SkipSeparation bool

	// Contacts marks which entities it collided with in the current frame
	Contacts map[int]bool

	CapsuleCollider     *collider.Capsule
	TriMeshCollider     *collider.TriMesh
	BoundingBoxCollider *collider.BoundingBox

	// stores the transformed collider (e.g. if the entity moves)
	TransformedCapsuleCollider     *collider.Capsule
	TransformedTriMeshCollider     *collider.TriMesh
	TransformedBoundingBoxCollider *collider.BoundingBox
}

func (c *ColliderComponent) AddToComponentContainer(container *ComponentContainer) {
	container.ColliderComponent = c
}

func (c *ColliderComponent) ComponentFlag() int {
	return ComponentFlagCollider
}

func (c *ColliderComponent) Synchronized() bool {
	return false
}

func (c *ColliderComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *ColliderComponent) Serialize() []byte {
	panic("wat")
}
