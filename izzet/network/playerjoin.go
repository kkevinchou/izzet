package network

type PlayerJoinMessage struct {
	PlayerID int
}

func (m PlayerJoinMessage) Type() MessageType {
	return MsgTypePlayerJoin
}
