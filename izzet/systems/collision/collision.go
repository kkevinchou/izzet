package collision

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/netsync"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/utils"
)

const (
	// the maximum number of times a distinct entity can have their collision resolved
	// this presents the collision resolution phase to go on forever
	resolveCountMax = 10
)

type World interface {
	GetSingleton() *singleton.Singleton
	GetPlayerEntity() entities.Entity
	QueryEntity(componentFlags int) []entities.Entity
	GetPlayer() *player.Player
	GetEntityByID(id int) entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}

type CollisionSystem struct {
	*base.BaseSystem
	world World
}

func NewCollisionSystem(world World) *CollisionSystem {
	return &CollisionSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
	}
}

func (s *CollisionSystem) Update(delta time.Duration) {
	if utils.IsClient() {
		player := s.world.GetPlayerEntity()
		netsync.ResolveCollisionsForPlayer(player, s.world)
	} else {
		netsync.ResolveCollisions(s.world)
	}
}

func (s *CollisionSystem) Name() string {
	return "CollisionSystem"
}
