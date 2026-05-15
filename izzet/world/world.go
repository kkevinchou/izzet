package world

import (
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entity"
)

type GameWorld struct {
	entities map[int]*entity.Entity

	commandFrameCount int
	spatialPartition  *spatialpartition.SpatialPartition

	sortFrame      int
	sortedEntities []*entity.Entity
}

func New() *GameWorld {
	return NewWithEntities(nil)
}

func NewWithEntities(entities map[int]*entity.Entity) *GameWorld {
	g := &GameWorld{
		sortFrame:        -1,
		entities:         map[int]*entity.Entity{},
		spatialPartition: spatialpartition.NewSpatialPartition(50, 10),
	}
	for _, e := range entities {
		g.AddEntity(e)
	}
	return g
}
