package server

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *Server) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	g.handleSpatialPartition()

	g.handlePlayerConnections()
	for _, s := range g.systems {
		s.Update(delta, g.world)
	}
	g.replicator.Update(delta, g.world)
}

func (g *Server) handlePlayerConnections() {
	select {
	case connection := <-g.newConnections:
		g.world.QueueEvent(events.PlayerJoinEvent{
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
