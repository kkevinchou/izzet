package server

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/events"
)

func (g *Server) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	g.handleSpatialPartition()

	g.handlePlayerConnections()
	for _, s := range g.systems {
		start := time.Now()
		s.Update(delta, g.world)
		metricName := fmt.Sprintf("%s_runtime", s.Name())
		g.metricsRegistry.Inc(metricName, float64(time.Since(start).Milliseconds()))
	}
}

func (g *Server) handlePlayerConnections() {
	select {
	case connection := <-g.newConnections:
		g.eventManager.PlayerJoinTopic.Write(events.PlayerJoinEvent{
			PlayerID:   connection.PlayerID,
			Connection: connection.Connection,
		})
	default:
		return
	}
}

func (g *Server) handleSpatialPartition() {
	var spatialEntities []spatialpartition.Entity
	for _, entity := range g.world.Entities() {
		if !entity.HasBoundingBox() {
			continue
		}
		spatialEntities = append(spatialEntities, entity)
	}
	g.world.SpatialPartition().IndexEntities(spatialEntities)
}
