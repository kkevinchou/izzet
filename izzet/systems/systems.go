package systems

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type System interface {
	Update(time.Duration, GameWorld)
}

type GameWorld interface {
	Entities() []*entities.Entity
	GetEntityByID(int) *entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
	GetEvents() []events.Event
	QueueEvent(events.Event)
	ClearEventQueue()
	AddEntity(*entities.Entity)
}

type App interface {
	IsClient() bool
	IsServer() bool
	CommandFrame() int
	GetPlayer(playerID int) *network.Player
	GetPlayerEntity() *entities.Entity
	GetPlayerCamera() *entities.Entity
}
