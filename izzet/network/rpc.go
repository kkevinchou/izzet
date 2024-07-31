package network

import "github.com/go-gl/mathgl/mgl64"

type RPCMessage struct {
	Pathfind *Pathfind
}

type Pathfind struct {
	Target mgl64.Vec3
}

func (m RPCMessage) Type() MessageType {
	return MsgTypeRPC
}
