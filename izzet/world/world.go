package world

import (
	"github.com/kkevinchou/izzet/internal/physics"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entity"
)

type GameWorld struct {
	entities map[int]*entity.Entity

	commandFrameCount int
	spatialPartition  *spatialpartition.SpatialPartition
	physicsWorld      *physics.World

	sortedEntities []*entity.Entity
}

func New() *GameWorld {
	return NewWithEntities(nil)
}

func NewWithEntities(entities map[int]*entity.Entity) *GameWorld {
	g := &GameWorld{
		entities:         map[int]*entity.Entity{},
		spatialPartition: spatialpartition.NewSpatialPartition(50, 10),
		physicsWorld:     physics.NewWorld(),
	}
	for _, e := range entities {
		g.AddEntity(e)
	}
	return g
}
