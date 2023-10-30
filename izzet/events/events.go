package events

import (
	"net"

	"github.com/kkevinchou/izzet/izzet/entities"
)

type Event interface{}

type PlayerJoinEvent struct {
	PlayerID       int
	Connection     net.Conn
	PlayerEntityID int
	PlayerCameraID int
}

type EntitySpawnEvent struct {
	Entity *entities.Entity
}

type PlayerDisconnectEvent struct {
	PlayerID int
}
