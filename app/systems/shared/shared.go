package shared

import (
	"github.com/kkevinchou/izzet/app/entities"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type GameWorld interface {
	GetEntityByID(int) *entities.Entity
	Entities() []*entities.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
