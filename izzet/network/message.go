package network

import (
	"time"
)

type MessageType int

const (
	MsgTypeAcceptConnection MessageType = iota
	MsgTypeGameStateUpdate
	MsgTypePlayerInput
	MsgTypeCreateEntity
	MsgTypeAckPlayerJoin
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
