package clientsystems

import (
	"net"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
)

type App interface {
	ModelLibrary() *modellibrary.ModelLibrary
	NetworkMessagesChannel() chan network.Message
	GetPlayerID() int
	CommandFrame() int
	IsConnected() bool
	GetPlayerConnection() net.Conn
	SetPlayerEntity(entity *entities.Entity)
	SetPlayerCamera(entity *entities.Entity)
	GetPlayerCamera() *entities.Entity
}
