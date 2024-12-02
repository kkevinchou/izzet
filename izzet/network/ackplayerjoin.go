package network

type AckPlayerJoinMessage struct {
	PlayerID    int
	EntityBytes []byte
	CameraBytes []byte
	Snapshot    []byte
	ProjectName string
}

func (m AckPlayerJoinMessage) Type() MessageType {
	return MsgTypeAckPlayerJoin
}
