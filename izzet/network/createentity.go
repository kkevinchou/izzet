package network

type CreateEntityMessage struct {
	OwnerID     int
	EntityBytes []byte
}

func (m CreateEntityMessage) Type() MessageType {
	return MsgTypeCreateEntity
}
