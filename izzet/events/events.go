package events

import (
	"net"

	"github.com/kkevinchou/izzet/izzet/entity"
)

type PlayerJoinEvent struct {
	PlayerID   int
	Connection net.Conn
}

type PlayerDisconnectEvent struct {
	PlayerID int
}

type EntitySpawnEvent struct {
	Entity *entity.Entity
}

type DestroyEntityEvent struct {
	EntityID int
}
