package types

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/lib/input"
)

type Item interface {
	Positionable
	IDable
	Ownable
}

type ItemReceiver interface {
	Give(item Item) error
}

type ItemGiver interface {
	Positionable
	Take(item Item) error
}

type ItemGiverReceiver interface {
	ItemGiver
	ItemReceiver
}

type Positionable interface {
	Position() mgl64.Vec3
	SetPosition(position mgl64.Vec3)
}

type IDable interface {
	ID() int
}

type Ownable interface {
	SetOwner(owner ItemReceiver)
	OwnedBy() ItemReceiver
	Owned() bool
}

type Worker interface {
	ItemGiverReceiver
	SetTarget(mgl64.Vec3)
	Velocity() mgl64.Vec3
	Heading() mgl64.Vec3
}

type Singleton interface {
	GetInput() input.Input
	SetInput(i input.Input)
}
