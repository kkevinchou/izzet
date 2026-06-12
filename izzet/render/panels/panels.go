package panels

import (
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entity"
)

const (
	propertiesWidth float32 = 450
)

type RenderContext interface {
	Width() int
	Height() int
	AspectRatio() float64
}

type GameWorld interface {
	Entities() []*entity.Entity
	AddEntity(entity *entity.Entity)
	GetEntityByID(id int) *entity.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
