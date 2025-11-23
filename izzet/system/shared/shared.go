package shared

import (
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entity"
)

type GameWorld interface {
	GetEntityByID(int) *entity.Entity
	Entities() []*entity.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
