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

func SendMessage(conn net.Conn, messageType MessageType, body any, frame int) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	message := Message{
		MessageType:  messageType,
		Timestamp:    time.Now(),
		Body:         bytes,
		CommandFrame: frame,
	}

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(message)
	if err != nil {
		return err
	}
	return nil
}
