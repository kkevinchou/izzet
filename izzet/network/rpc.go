package network

import "github.com/go-gl/mathgl/mgl32"

type RPCMessage struct {
	Pathfind *Pathfind
}

type Pathfind struct {
	Goal mgl32.Vec3
}

func (m RPCMessage) Type() MessageType {
	return MsgTypeRPC
}
