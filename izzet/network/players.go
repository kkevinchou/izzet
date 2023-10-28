package network

import (
	"net"
)

type Player struct {
	ID                         int
	Connection                 net.Conn
	InMessageChannel           chan Message
	OutMessageChannel          chan Message
	DisconnectChannel          chan bool
	LastInputLocalCommandFrame int // local command frame from the client
}
