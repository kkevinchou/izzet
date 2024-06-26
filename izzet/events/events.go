package events

import (
	"net"

	"github.com/kkevinchou/izzet/izzet/entities"
)

type PlayerJoinEvent struct {
	PlayerID   int
	Connection net.Conn
}

type PlayerDisconnectEvent struct {
	PlayerID int
}

type EntitySpawnEvent struct {
	Entity *entities.Entity
}

type DestroyEntityEvent struct {
	EntityID int
}
