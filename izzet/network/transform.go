package network

import "github.com/go-gl/mathgl/mgl64"

type Transform struct {
	EntityID    int
	Position    mgl64.Vec3
	Orientation mgl64.Quat
	Velocity    mgl64.Vec3
	Animation   string
}

type GameStateUpdateMessage struct {
	Transforms            []Transform
	LastInputCommandFrame int
}

func (m GameStateUpdateMessage) Type() MessageType {
	return MsgTypeGameStateUpdate
}
