package network

import "github.com/go-gl/mathgl/mgl64"

type EntityState struct {
	EntityID  int
	Position  mgl64.Vec3
	Rotation  mgl64.Quat
	Velocity  mgl64.Vec3
	Animation string
}

type GameStateUpdateMessage struct {
	EntityStates          []EntityState
	LastInputCommandFrame int
	GlobalCommandFrame    int
}

func (m GameStateUpdateMessage) Type() MessageType {
	return MsgTypeGameStateUpdate
}
