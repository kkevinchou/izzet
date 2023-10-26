package network

import (
	"time"
)

type MessageType int

const (
	MsgTypeAcceptConnection MessageType = iota
	MsgTypeAckCreatePlayer
	MsgTypeGameStateUpdate
	MsgTypePlayerInput
)

type Message struct {
	SenderID     int
	MessageType  MessageType
	CommandFrame int
	Timestamp    time.Time

	Body []byte
}

type AcceptMessage struct {
	ID int
}
