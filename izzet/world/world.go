package world

import (
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type GameWorld struct {
	entities map[int]*entities.Entity

	commandFrameCount int
	spatialPartition  *spatialpartition.SpatialPartition

	sortFrame      int
	sortedEntities []*entities.Entity
}

func New() *GameWorld {
	return NewWithEntities(map[int]*entities.Entity{})
}

func NewWithEntities(entities map[int]*entities.Entity) *GameWorld {
	g := &GameWorld{
		sortFrame:        -1,
		entities:         entities,
		spatialPartition: spatialpartition.NewSpatialPartition(50, 10),
	}
	return g
}
