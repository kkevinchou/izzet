package network

import (
	"net"
)

type Player struct {
	ID                int
	Connection        net.Conn
	InMessageChannel  chan Message
	OutMessageChannel chan Message
}
