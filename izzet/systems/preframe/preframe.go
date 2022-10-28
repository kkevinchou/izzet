package preframe

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/systems/base"
)

type World interface {
	GetSingleton() *singleton.Singleton
	QueryEntity(componentFlags int) []entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
	GetPlayerEntity() entities.Entity
	GetEntityByID(id int) entities.Entity
}

type PreFrameSystem struct {
	*base.BaseSystem

	world World
}

func NewPreFrameSystem(world World) *PreFrameSystem {
	return &PreFrameSystem{
		world: world,
	}
}

func (s *PreFrameSystem) Update(delta time.Duration) {
	s.world.SpatialPartition().FrameSetup(s.world)

}

func (s *PreFrameSystem) Name() string {
	return "PreFrameSystem"
}
