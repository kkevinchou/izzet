package network

import "github.com/go-gl/mathgl/mgl64"

type Transform struct {
	EntityID    int
	Position    mgl64.Vec3
	Orientation mgl64.Quat
}

type GameStateUpdateMessage struct {
	Transforms            []Transform
	LastInputCommandFrame int
}
