package clientsystems

import "github.com/kkevinchou/izzet/izzet/network"

type App interface {
	NetworkMessagesChannel() chan network.Message
	GetPlayerID() int
}
