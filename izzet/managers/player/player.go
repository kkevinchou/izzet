package player

import (
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/lib/network"
)

type World interface {
	CommandFrame() int
}

type Player struct {
	ID       int
	EntityID int
	Client   types.NetworkClient

	LastInputLocalCommandFrame  int // the player's last command frame
	LastInputGlobalCommandFrame int // the gcf when this input was received

	lastNetworkPullCommandFrame    int
	lastNetworkPullNetworkMessages []*network.Message
	world                          World
}

// NetworkMessages pulls network messages for the player. This function caches network
// messages for the same command frame
func (p *Player) NetworkMessages() []*network.Message {
	commandFrame := p.world.CommandFrame()
	if p.lastNetworkPullCommandFrame != commandFrame {
		p.lastNetworkPullNetworkMessages = p.Client.PullIncomingMessages()
		p.lastNetworkPullCommandFrame = commandFrame
	}

	return p.lastNetworkPullNetworkMessages
}

type PlayerManager struct {
	players   []*Player
	playerMap map[int]*Player
	world     World
}

func NewPlayerManager(world World) *PlayerManager {
	return &PlayerManager{
		players:   []*Player{},
		playerMap: map[int]*Player{},
		world:     world,
	}
}

func (p *PlayerManager) RegisterPlayer(playerID int, client types.NetworkClient) {
	player := &Player{ID: playerID, Client: client, world: p.world}
	p.playerMap[playerID] = player
	p.players = append(p.players, player)
}

func (p *PlayerManager) GetPlayer(id int) *Player {
	return p.playerMap[id]
}

func (p *PlayerManager) GetPlayers() []*Player {
	return p.players
}
