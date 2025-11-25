package network

type AckPlayerJoinMessage struct {
	ProjectName     string
	PlayerID        int
	PlayerEntityID  int
	CameraEntityID  int
	SerializedWorld []byte
}

func (m AckPlayerJoinMessage) Type() MessageType {
	return MsgTypeAckPlayerJoin
}
