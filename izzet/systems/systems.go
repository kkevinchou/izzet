package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type System interface {
	Update(time.Duration, GameWorld)
	Name() string
}

type GameWorld interface {
	Entities() []*entities.Entity
	GetEntityByID(int) *entities.Entity
	DeleteEntity(int)
	SpatialPartition() *spatialpartition.SpatialPartition
	AddEntity(*entities.Entity)
}

type App interface {
	IsClient() bool
	IsServer() bool
	CommandFrame() int
	GetPlayer(playerID int) *network.Player
	GetPlayerEntity() *entities.Entity
	GetPlayerCamera() *entities.Entity
	MetricsRegistry() *metrics.MetricsRegistry
	CollisionObserver() *collisionobserver.CollisionObserver
	World() *world.GameWorld
}
