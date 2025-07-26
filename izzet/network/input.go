package network

import "github.com/kkevinchou/izzet/internal/input"

type InputMessage struct {
	Input input.Input
}

func (m InputMessage) Type() MessageType {
	return MsgTypePlayerInput
}
