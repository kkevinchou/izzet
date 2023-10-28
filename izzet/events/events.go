package events

import "net"

type Event interface{}

type PlayerJoinEvent struct {
	PlayerID       int
	Connection     net.Conn
	PlayerEntityID int
	PlayerCameraID int
}

type PlayerDisconnectEvent struct {
	PlayerID int
}
