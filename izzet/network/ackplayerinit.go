package network

type AckPlayerJoinMessage struct {
	PlayerID    int
	EntityBytes []byte
	CameraBytes []byte
}

func (m AckPlayerJoinMessage) Type() MessageType {
	return MsgTypeAckPlayerJoin
}
