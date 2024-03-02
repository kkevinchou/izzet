package serversystems

import (
	"net"

	"github.com/kkevinchou/izzet/app/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/kitolib/input"
	"github.com/kkevinchou/kitolib/metrics"
)

type App interface {
	GetPlayers() map[int]*network.Player
	RegisterPlayer(playerID int, connection net.Conn) *network.Player
	InputBuffer() *inputbuffer.InputBuffer
	CommandFrame() int
	ModelLibrary() *modellibrary.ModelLibrary
	GetPlayer(playerID int) *network.Player
	GetPlayerInput(playerID int) input.Input
	SetPlayerInput(playerID int, input input.Input)
	DeregisterPlayer(playerID int)
	SerializeWorld() []byte
	MetricsRegistry() *metrics.MetricsRegistry
	EventsManager() *events.EventManager
	SystemNames() []string
}
