package network

func NewBaseMessage(senderID int, messageType MessageType, commandFrame int) Message {
	return Message{
		SenderID:     senderID,
		MessageType:  messageType,
		CommandFrame: commandFrame,
	}
}
