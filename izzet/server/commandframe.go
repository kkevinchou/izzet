package server

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *Server) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	g.handleSpatialPartition()

	g.handlePlayerConnections()
	g.handleNetworkMessages()
	for _, s := range g.systems {
		s.Update(delta, g.world)
	}
	g.replicator.Update(delta, g.world)
}

func (g *Server) handleNetworkMessages() {
	for _, player := range g.players {
		decoder := json.NewDecoder(player.Connection)
		var message network.Message
		err := decoder.Decode(&message)
		if err != nil {
			reader := decoder.Buffered()
			bytes, err2 := io.ReadAll(reader)
			if err2 != nil {
				fmt.Println(fmt.Errorf("error reading remaining bytes %w", err2))
			}
			fmt.Println(fmt.Errorf("error decoding message from player %w remaining buffered data: %s", err, string(bytes)))
			continue
		}

		if message.MessageType == network.MsgTypePlayerInput {
			player.InMessageChannel <- message
		}
	}
}

func (g *Server) handlePlayerConnections() {
	select {
	case connection := <-g.newConnections:
		g.world.QueueEvent(events.PlayerJoinEvent{PlayerID: connection.PlayerID, Connection: connection.Connection})
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
