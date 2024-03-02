package clientsystems

import (
	"net"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serverstats"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

type App interface {
	ModelLibrary() *modellibrary.ModelLibrary
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
	MetricsRegistry() *metrics.MetricsRegistry
	Client() network.IzzetClient
	StateBuffer() *StateBuffer
	GetFrameInput() input.Input
	GetFrameInputPtr() *input.Input
	SetServerStats(stats serverstats.ServerStats)
}
