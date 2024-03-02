package shared

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type GameWorld interface {
	GetEntityByID(int) *entities.Entity
	Entities() []*entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
