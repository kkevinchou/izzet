package world

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type GameWorld struct {
	// gameOver bool
	// camera *camera.Camera

	entities map[int]*entities.Entity

	// editHistory *edithistory.EditHistory

	commandFrameCount int
	spatialPartition  *spatialpartition.SpatialPartition
	// relativeMouseOrigin [2]int32
	// relativeMouseActive bool

	// navigationMesh  *navmesh.NavigationMesh
	// metricsRegistry *metrics.MetricsRegistry

	sortFrame      int
	sortedEntities []*entities.Entity
	frameInput     input.Input
}

func New(entities map[int]*entities.Entity) *GameWorld {
	g := &GameWorld{
		sortFrame:        -1,
		entities:         entities,
		spatialPartition: spatialpartition.NewSpatialPartition(200, 25),
	}
	return g
}
