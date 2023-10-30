package events

import (
	"net"

	"github.com/kkevinchou/izzet/izzet/entities"
)

type Event interface{}

type PlayerJoinEvent struct {
	PlayerID   int
	Connection net.Conn
}

type EntitySpawnEvent struct {
	Entity *entities.Entity
}

type PlayerDisconnectEvent struct {
	PlayerID int
}
