package network

import (
	"encoding/json"
	"net"
	"time"
)

func NewBaseMessage(senderID int, messageType MessageType, commandFrame int) Message {
	return Message{
		SenderID:     senderID,
		MessageType:  messageType,
		CommandFrame: commandFrame,
	}
}

func SendMessage(conn net.Conn, messageType MessageType, body any, frame int) {
	bytes, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	message := Message{
		MessageType:  messageType,
		Timestamp:    time.Now(),
		Body:         bytes,
		CommandFrame: frame,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(messageBytes)
	if err != nil {
		panic(err)
	}
}
