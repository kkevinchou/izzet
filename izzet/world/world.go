package world

import (
	"github.com/kkevinchou/izzet/izzet/camera"
	"github.com/kkevinchou/izzet/izzet/edithistory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/kitolib/metrics"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

type GameWorld struct {
	gameOver bool

	camera *camera.Camera

	entities map[int]*entities.Entity

	editHistory *edithistory.EditHistory

	commandFrameCount   int
	spatialPartition    *spatialpartition.SpatialPartition
	relativeMouseOrigin [2]int32
	relativeMouseActive bool

	navigationMesh  *navmesh.NavigationMesh
	metricsRegistry *metrics.MetricsRegistry

	sortFrame      int
	sortedEntities []*entities.Entity
}

func New() *GameWorld {
	g := &GameWorld{}
	return g
}
