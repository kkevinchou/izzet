package systems

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type System interface {
	Update(time.Duration, GameWorld)
}

type GameWorld interface {
	Entities() []*entities.Entity
	GetEntityByID(int) *entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
	GetFrameInput() input.Input
	SetInputCameraOrientation(mgl64.Quat)
	GetEvents() []events.Event
	QueueEvent(events.Event)
	ClearEventQueue()
	AddEntity(*entities.Entity)
}
