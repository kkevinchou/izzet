package observers

// OnSpatialQuery(entityID int, count int)
// OnCollisionCheck(entityID int)
// OnCollisionResolution(entityID int)
type PhysicsObserver struct {
	SpatialQuery        map[int]int
	CollisionCheck      map[int]int
	CollisionResolution map[int]int
}

func NewPhysicsObserver() *PhysicsObserver {
	return &PhysicsObserver{
		SpatialQuery:        map[int]int{},
		CollisionCheck:      map[int]int{},
		CollisionResolution: map[int]int{},
	}
}

func (o *PhysicsObserver) OnSpatialQuery(entityID int, count int) {
	o.SpatialQuery[entityID] += count
}
func (o *PhysicsObserver) OnCollisionCheck(entityID int) {
	o.CollisionCheck[entityID] += 1
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
}
