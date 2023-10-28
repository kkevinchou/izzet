package network

import "github.com/kkevinchou/kitolib/input"

type InputMessage struct {
	Input input.Input
}

func (m InputMessage) Type() MessageType {
	return MsgTypePlayerInput
}
