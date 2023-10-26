package network

import "github.com/go-gl/mathgl/mgl64"

type Transform struct {
	EntityID int
	Position mgl64.Vec3
}

type GameStateUpdateMessage struct {
	Transforms []Transform
}
