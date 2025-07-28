package network

import "github.com/go-gl/mathgl/mgl64"

type RPCMessage struct {
	Pathfind     *Pathfind
	CreateEntity *CreateEntityRPC
}

type Pathfind struct {
	Goal mgl64.Vec3
}

type CreateEntityRPC struct {
	EntityType string
	Patrol     bool
	// Position   mgl64.Vec3
}

func (m RPCMessage) Type() MessageType {
	return MsgTypeRPC
}
