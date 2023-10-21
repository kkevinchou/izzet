package observers

import (
	"github.com/kkevinchou/izzet/izzet/app"
	"github.com/kkevinchou/izzet/izzet/entities"
)

// OnSpatialQuery(entityID int, count int)
// OnCollisionCheck(entityID int)
// OnCollisionResolution(entityID int)
type PhysicsObserver struct {
	SpatialQuery        map[int]int
	CollisionCheck      map[int]int
	CollisionResolution map[int]int

	CollisionCheckTriangle map[int]int
	CollisionCheckTriMesh  map[int]int
	CollisionCheckCapsule  map[int]int

	BoundingBoxCheck map[int]int
}

func NewPhysicsObserver() *PhysicsObserver {
	return &PhysicsObserver{
		SpatialQuery:           map[int]int{},
		CollisionCheck:         map[int]int{},
		CollisionResolution:    map[int]int{},
		CollisionCheckTriangle: map[int]int{},
		CollisionCheckTriMesh:  map[int]int{},
		CollisionCheckCapsule:  map[int]int{},
		BoundingBoxCheck:       map[int]int{},
	}
}
func (o *PhysicsObserver) OnBoundingBoxCheck(e1 *entities.Entity, e2 *entities.Entity) {
	o.BoundingBoxCheck[e1.GetID()] += 1
}
func (o *PhysicsObserver) OnSpatialQuery(entityID int, count int) {
	o.SpatialQuery[entityID] += count
}
func (o *PhysicsObserver) OnCollisionCheck(e1 *entities.Entity, e2 *entities.Entity) {
	o.CollisionCheck[e1.GetID()] += 1
	if app.IsCapsuleCapsuleCollision(e1, e2) {
		o.CollisionCheckCapsule[e1.GetID()] += 1
	} else if ok, _, _ := app.IsCapsuleTriMeshCollision(e1, e2); ok {
		o.CollisionCheckTriMesh[e1.GetID()] += 1
		o.CollisionCheckTriangle[e1.GetID()] += len(e2.Collider.TriMeshCollider.Triangles)
	}
}
func (o *PhysicsObserver) OnCollisionResolution(entityID int) {
	o.CollisionResolution[entityID] += 1
}
func (o *PhysicsObserver) Clear() {
	for k := range o.SpatialQuery {
		o.SpatialQuery[k] = 0
	}
	for k := range o.CollisionCheck {
		o.CollisionCheck[k] = 0
	}
	for k := range o.CollisionResolution {
		o.CollisionResolution[k] = 0
	}
	for k := range o.CollisionCheckTriMesh {
		o.CollisionCheckTriMesh[k] = 0
	}
	for k := range o.CollisionCheckTriangle {
		o.CollisionCheckTriangle[k] = 0
	}
	for k := range o.CollisionCheckCapsule {
		o.CollisionCheckCapsule[k] = 0
	}
	for k := range o.BoundingBoxCheck {
		o.BoundingBoxCheck[k] = 0
	}
}
