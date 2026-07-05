package server

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/event"
	"github.com/kkevinchou/izzet/izzet/telemetry"
)

func (g *Server) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	g.world.ReindexSpatialEntities()
	g.handlePlayerConnections()
	for _, s := range g.systems {
		start := time.Now()
		s.Update(delta, g.world)
		metricName := fmt.Sprintf("%s_runtime", s.Name())
		telemetry.ServerRegistry().Inc(metricName, float64(time.Since(start).Milliseconds()))
	}
}

func (g *Server) handlePlayerConnections() {
	select {
	case connection := <-g.newConnections:
		g.eventManager.PlayerJoinTopic.Write(event.PlayerJoinEvent{
			PlayerID:   connection.PlayerID,
			Connection: connection.Connection,
		})
	default:
		return
	}
}
