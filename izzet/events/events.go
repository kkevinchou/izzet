package events

import "github.com/kkevinchou/izzet/izzet/playercommand/protogen/playercommand"

type EventType string

var EventTypeUnregisterEntity EventType = "UNREGISTER"
var EventTypePlayerCommand EventType = "PLAYERCOMMAND"
var EventTypeConsoleEnabled EventType = "CONSOLE_ENABLED"
var EventTypeRPC EventType = "RPC"

type Event interface {
	Type() EventType
}

type PlayerCommandEvent struct {
	Command *playercommand.Wrapper
}

func (e *PlayerCommandEvent) Type() EventType {
	return EventTypePlayerCommand
}

type UnregisterEntityEvent struct {
	EntityID           int `json:"entity_id"`
	GlobalCommandFrame int `json:"global_command_frame"`
}

func (e *UnregisterEntityEvent) Type() EventType {
	return EventTypeUnregisterEntity
}

type ConsoleEnabledEvent struct {
}

func (e *ConsoleEnabledEvent) Type() EventType {
	return EventTypeConsoleEnabled
}

type RPCEvent struct {
	PlayerID int
	Command  string
}

func (e *RPCEvent) Type() EventType {
	return EventTypeRPC
}
