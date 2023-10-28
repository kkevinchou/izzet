package clientsystems

import (
	"net"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems/clientsystems/commandframe"
	"github.com/kkevinchou/kitolib/metrics"
)

type App interface {
	ModelLibrary() *modellibrary.ModelLibrary
	NetworkMessagesChannel() chan network.Message
	GetPlayerID() int
	CommandFrame() int
	IsConnected() bool
	GetPlayerConnection() net.Conn
	GetPlayerEntity() *entities.Entity
	GetPlayerCamera() *entities.Entity
	GetCommandFrameHistory() *commandframe.CommandFrameHistory
	MetricsRegistry() *metrics.MetricsRegistry
}
