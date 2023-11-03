package network

type PingMessage struct {
	UnixTime int64
}

func (m PingMessage) Type() MessageType {
	return MsgTypePing
}
