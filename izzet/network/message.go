package network

import (
	"time"
)

type MessageType int

const (
	MsgTypeAcceptConnection MessageType = iota
	MsgTypeAckCreatePlayer
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
