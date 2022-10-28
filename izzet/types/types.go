package types

import "github.com/kkevinchou/izzet/lib/network"

type MovementType int

const (
	MovementTypeSteering    MovementType = iota
	MovementTypeDirectional MovementType = iota
)

type GameMode string

const (
	GameModeEditor  GameMode = "EDITOR"
	GameModePlaying GameMode = "PLAYING"
)

type NetworkClient interface {
	SendMessage(messageType int, messageBody any) error
	PullIncomingMessages() []*network.Message
}

type Window string

const (
	WindowConsole   Window = "CONSOLE"
	WindowGame      Window = "GAME"
	WindowDebug     Window = "DEBUG"
	WindowInventory Window = "INVENTORY"
)
