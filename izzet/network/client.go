package network

import (
	"encoding/json"
	"net"
	"time"
)

type connectionImpl struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
}

type IzzetClient interface {
	Send(messageBody Message, frame int) error
	Recv() (MessageTransport, error)
}

func NewClient(conn net.Conn) IzzetClient {
	return &connectionImpl{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}
}

func (c *connectionImpl) Send(message Message, frame int) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	messageTransport := MessageTransport{
		MessageType:  message.Type(),
		Timestamp:    time.Now(),
		Body:         bytes,
		CommandFrame: frame,
	}

	err = c.encoder.Encode(messageTransport)
	if err != nil {
		return err
	}
	return nil
}

func (c *connectionImpl) Recv() (MessageTransport, error) {
	var message MessageTransport
	err := c.decoder.Decode(&message)
	if err != nil {
		return MessageTransport{}, err
	}

	return message, nil
}
