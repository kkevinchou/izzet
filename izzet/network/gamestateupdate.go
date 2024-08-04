package network

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/serverstats"
)

type EntityState struct {
	EntityID       int
	Position       mgl32.Vec3
	Rotation       mgl32.Quat
	Velocity       mgl32.Vec3
	GravityEnabled bool
	Animation      string
}

type GameStateUpdateMessage struct {
	EntityStates          []EntityState
	LastInputCommandFrame int
	GlobalCommandFrame    int
	ServerStats           serverstats.ServerStats
	DestroyedEntities     []int
}

func (m GameStateUpdateMessage) Type() MessageType {
	return MsgTypeGameStateUpdate
}
