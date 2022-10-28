package networkdispatch

import (
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/lib/network"
)

func connectedPlayersMessageFetcher(world World) []*network.Message {
	playerManager := directory.GetDirectory().PlayerManager()
	var allMessages []*network.Message

	for _, player := range playerManager.GetPlayers() {
		messages := player.Client.PullIncomingMessages()
		allMessages = append(allMessages, messages...)
	}

	return allMessages
}

func clientMessageFetcher(world World) []*network.Message {
	player := world.GetPlayer()
	return player.NetworkMessages()
}
