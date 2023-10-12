package world

import (
	"github.com/kkevinchou/izzet/izzet/entities"
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
}

func New() *GameWorld {
	g := &GameWorld{
		entities:         map[int]*entities.Entity{},
		spatialPartition: spatialpartition.NewSpatialPartition(200, 25),
	}
	return g
}
