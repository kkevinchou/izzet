package server

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/panels"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/metrics"
)

func (g *Server) MetricsRegistry() *metrics.MetricsRegistry {
	return g.metricsRegistry
}

func (g *Server) LoadWorld(name string) bool {
	if name == "" {
		return false
	}

	filename := fmt.Sprintf("./%s.json", name)
	world, err := g.serializer.ReadFromFile(filename)
	if err != nil {
		fmt.Println("failed to load world", filename, err)
		panic(err)
	}

	g.editHistory.Clear()
	g.world.SpatialPartition().Clear()

	var maxID int
	for _, e := range world.Entities() {
		if e.ID > maxID {
			maxID = e.ID
		}
		g.entities[e.ID] = e
	}

	if len(g.entities) > 0 {
		entities.SetNextID(maxID + 1)
	}

	panels.SelectEntity(nil)
	g.SetWorld(world)
	return true
}

func (g *Server) SetWorld(world *world.GameWorld) {
	g.world = world
}

func (g *Server) ModelLibrary() *modellibrary.ModelLibrary {
	return g.modelLibrary
}
