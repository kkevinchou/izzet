package knetwork

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/kitolib/input"
)

const (
	MessageTypeAcceptConnection int = iota
	MessageTypeAckCreatePlayer
	MessageTypeInput
	MessageTypeCreatePlayer
	MessageTypeGameStateUpdate
	MessageTypePing
	MessageTypeAckPing
	MessageTypeRPC
)

type AcceptMessage struct {
	ID int
}

type CreatePlayerMessage struct {
}

type AckCreatePlayerMessage struct {
	PlayerID    int
	EntityID    int
	CameraID    int
	Position    mgl64.Vec3
	Orientation mgl64.Quat

	Entities map[int]EntitySnapshot
}

type EntitySnapshot struct {
	ID   int
	Type int

	// Movement
	Position    mgl64.Vec3
	Orientation mgl64.Quat
	Velocity    mgl64.Vec3

	Animation string

	Components map[int][]byte // protobuf
}

type Event struct {
	Type  events.EventType
	Bytes []byte
}

type EventI interface {
	TypeAsInt() int
	Serialize() ([]byte, error)
}

type GameStateUpdateMessage struct {
	ServerStats                 map[string]string
	LastInputCommandFrame       int
	LastInputGlobalCommandFrame int
	CurrentGlobalCommandFrame   int
	Entities                    map[int]EntitySnapshot
	Events                      []Event
}

type InputMessage struct {
	PlayerCommands []byte // protobuf
	CommandFrame   int
	Input          input.Input
}

type PingMessage struct {
	SendTime time.Time
}

type AckPingMessage struct {
	PingSendTime time.Time
}

type RPCMessage struct {
	Command string
}
