package shared

import (
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type GameWorld interface {
	GetEntityByID(int) *entities.Entity
	GetKinematicEntityByID(int) *entities.Entity
	Entities() []*entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
