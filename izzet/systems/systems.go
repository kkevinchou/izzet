package systems

import (
	"time"

	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/collisionobserver"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/runtimeconfig"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/world"
)

type System interface {
	Update(time.Duration, GameWorld)
	Name() string
}

type GameWorld interface {
	Entities() []*entity.Entity
	GetEntityByID(int) *entity.Entity
	DeleteEntity(int)
	SpatialPartition() *spatialpartition.SpatialPartition
	AddEntity(*entity.Entity)
}

type App interface {
	IsClient() bool
	IsServer() bool
	CommandFrame() int
	GetPlayer(playerID int) *network.Player
	GetPlayerEntity() *entity.Entity
	GetPlayerCamera() *entity.Entity
	CollisionObserver() *collisionobserver.CollisionObserver
	World() *world.GameWorld
	AppMode() types.AppMode
	RuntimeConfig() *runtimeconfig.RuntimeConfig
	PredictionDebugLogging() bool
}
