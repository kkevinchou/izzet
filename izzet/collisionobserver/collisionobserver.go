package collisionobserver

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/physicsutils"
)

// OnSpatialQuery(entityID int, count int)
// OnCollisionCheck(entityID int)
// OnCollisionResolution(entityID int)
type CollisionObserver struct {
	SpatialQuery        map[int]int
	CollisionCheck      map[int]int
	CollisionResolution map[int]int

	CollisionCheckTriangle map[int]int
	CollisionCheckTriMesh  map[int]int
	CollisionCheckCapsule  map[int]int

	BoundingBoxCheck map[int]int
}

func NewCollisionObserver() *CollisionObserver {
	return &CollisionObserver{
		SpatialQuery:           map[int]int{},
		CollisionCheck:         map[int]int{},
		CollisionResolution:    map[int]int{},
		CollisionCheckTriangle: map[int]int{},
		CollisionCheckTriMesh:  map[int]int{},
		CollisionCheckCapsule:  map[int]int{},
		BoundingBoxCheck:       map[int]int{},
	}
}
func (o *CollisionObserver) OnBoundingBoxCheck(e1 *entities.Entity, e2 *entities.Entity) {
	o.BoundingBoxCheck[e1.GetID()] += 1
}
func (o *CollisionObserver) OnSpatialQuery(entityID int, count int) {
	o.SpatialQuery[entityID] += count
}
func (o *CollisionObserver) OnCollisionCheck(e1 *entities.Entity, e2 *entities.Entity) {
	o.CollisionCheck[e1.GetID()] += 1
	if physicsutils.IsCapsuleCapsuleCollision(e1, e2) {
		o.CollisionCheckCapsule[e1.GetID()] += 1
	} else if ok, _, _ := physicsutils.IsCapsuleTriMeshCollision(e1, e2); ok {
		o.CollisionCheckTriMesh[e1.GetID()] += 1
		o.CollisionCheckTriangle[e1.GetID()] += len(e2.Collider.TriMeshCollider.Triangles)
	}
}
func (o *CollisionObserver) OnCollisionResolution(entityID int) {
	o.CollisionResolution[entityID] += 1
}
func (o *CollisionObserver) Clear() {
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

var NullCollisionExplorer nullCollisionObserverType

type nullCollisionObserverType struct {
}

func (o nullCollisionObserverType) OnBoundingBoxCheck(e1 *entities.Entity, e2 *entities.Entity) {
}
func (o nullCollisionObserverType) OnSpatialQuery(entityID int, count int) {
}
func (o nullCollisionObserverType) OnCollisionCheck(e1 *entities.Entity, e2 *entities.Entity) {
}
func (o nullCollisionObserverType) OnCollisionResolution(entityID int) {
}
func (o nullCollisionObserverType) Clear() {
}
