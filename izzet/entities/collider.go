package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/types"
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
	proxyCapsuleCollider     *ProxyCapsule     `json:"-"`
	proxyTriMeshCollider     *ProxyTriMesh     `json:"-"`
	proxyBoundingBoxCollider *ProxyBoundingBox `json:"-"`
}

type ProxyCapsule struct {
	collider.Capsule
	Dirty bool
}

type ProxyTriMesh struct {
	collider.TriMesh
	Dirty bool
}

type ProxyBoundingBox struct {
	collider.BoundingBox
	Dirty bool
}

func (c *ColliderComponent) ProxyCapsule(transform mgl64.Mat4) collider.Capsule {
	if c.proxyCapsuleCollider.Dirty {
		c.proxyCapsuleCollider.Capsule = c.CapsuleCollider.Transform(transform)
		c.proxyCapsuleCollider.Dirty = false
	}
	return c.proxyCapsuleCollider.Capsule
}

func (c *ColliderComponent) ProxyTriMesh(transform mgl64.Mat4) collider.TriMesh {
	if c.proxyTriMeshCollider.Dirty {
		c.proxyTriMeshCollider.TriMesh = c.TriMeshCollider.Transform(transform)
		c.proxyTriMeshCollider.Dirty = false
	}
	return c.proxyTriMeshCollider.TriMesh
}

func (c *ColliderComponent) ProxyBoundingBox(transform mgl64.Mat4) collider.BoundingBox {
	if c.proxyBoundingBoxCollider.Dirty {
		c.proxyBoundingBoxCollider.BoundingBox = c.BoundingBoxCollider.Transform(transform)
		c.proxyBoundingBoxCollider.Dirty = false
	}
	return c.proxyBoundingBoxCollider.BoundingBox
}

func CreateCapsuleColliderComponent(colliderGroup, collisionMask types.ColliderGroupFlag, capsule collider.Capsule) *ColliderComponent {
	bb := collider.BoundingBox{
		MinVertex: capsule.Bottom.Sub(mgl64.Vec3{capsule.Radius, capsule.Radius, capsule.Radius}),
		MaxVertex: capsule.Top.Add(mgl64.Vec3{capsule.Radius, capsule.Radius, capsule.Radius}),
	}

	return &ColliderComponent{
		ColliderGroup:            colliderGroup,
		CollisionMask:            collisionMask,
		CapsuleCollider:          &capsule,
		proxyCapsuleCollider:     &ProxyCapsule{Capsule: capsule, Dirty: true},
		BoundingBoxCollider:      &bb,
		proxyBoundingBoxCollider: &ProxyBoundingBox{BoundingBox: bb, Dirty: true},
	}
}

func CreateTriMeshColliderComponent(colliderGroup, collisionMask types.ColliderGroupFlag, triMesh collider.TriMesh, boundingBox collider.BoundingBox) *ColliderComponent {
	return &ColliderComponent{
		ColliderGroup:            colliderGroup,
		CollisionMask:            collisionMask,
		TriMeshCollider:          &triMesh,
		proxyTriMeshCollider:     &ProxyTriMesh{TriMesh: triMesh, Dirty: true},
		BoundingBoxCollider:      &boundingBox,
		proxyBoundingBoxCollider: &ProxyBoundingBox{BoundingBox: boundingBox, Dirty: true},
	}
}

func (e *Entity) HasCapsuleCollider() bool {
	if e.Collider == nil {
		return false
	}
	return e.Collider.CapsuleCollider != nil
}

func (e *Entity) HasTriMeshCollider() bool {
	if e.Collider == nil {
		return false
	}
	return e.Collider.TriMeshCollider != nil
}

func (e *Entity) HasBoundingBox() bool {
	if e.Collider == nil {
		return false
	}
	return e.Collider.BoundingBoxCollider != nil
}

func (e *Entity) CapsuleCollider() collider.Capsule {
	return e.Collider.ProxyCapsule(WorldTransform(e))
}

func (e *Entity) TriMeshCollider() collider.TriMesh {
	return e.Collider.ProxyTriMesh(WorldTransform(e))
}

func (e *Entity) BoundingBox() collider.BoundingBox {
	return e.Collider.ProxyBoundingBox(WorldTransform(e))
}
