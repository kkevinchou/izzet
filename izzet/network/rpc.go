package network

import "github.com/go-gl/mathgl/mgl64"

type RPCMessage struct {
	Pathfind     *Pathfind
	CreateEntity *CreateEntityRPC
	RessurectRPC *RessurectRPC
}

type Pathfind struct {
	Goal mgl64.Vec3
}

type CreateEntityRPC struct {
	EntityType string
	Patrol     bool
}

type RessurectRPC struct {
	ID int
}

func (m RPCMessage) Type() MessageType {
	return MsgTypeRPC
}
