package clientsystems

import (
	"net"

	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/izzet/izzet/world"
)

type App interface {
	AssetManager() *assets.AssetManager
	NetworkMessagesChannel() chan network.MessageTransport
	GetPlayerID() int
	CommandFrame() int
	IsConnected() bool
	IsClient() bool
	IsServer() bool
	GetPlayerConnection() net.Conn
	GetPlayerEntity() *entities.Entity
	GetPlayerCamera() *entities.Entity
	GetCommandFrameHistory() *CommandFrameHistory
	Client() network.IzzetClient
	StateBuffer() *StateBuffer
	GetFrameInput() input.Input
	GetFrameInputPtr() *input.Input
	SetServerStats(stats serverstats.ServerStats)
	World() *world.GameWorld
	PredictionDebugLogging() bool
	SetPredictionDebugLogging(value bool)
}
