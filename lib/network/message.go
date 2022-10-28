package network

import (
	"time"
)

const (
	MessageTypeAcceptConnection int = iota
	MessageTypeAckCreatePlayer
)

type Message struct {
	SenderID     int
	MessageType  int
	CommandFrame int
	Timestamp    time.Time

	Body []byte
}

type AcceptMessage struct {
	ID int
}
