package events

import "net"

type Event interface{}

type PlayerJoinEvent struct {
	PlayerID   int
	Connection net.Conn
}
