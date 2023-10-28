package network

import (
	"net"
)

type Player struct {
	ID                         int
	Connection                 net.Conn
	InMessageChannel           chan MessageTransport
	OutMessageChannel          chan MessageTransport
	DisconnectChannel          chan bool
	LastInputLocalCommandFrame int // local command frame from the client
	Client                     IzzetClient
}
