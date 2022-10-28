package izzet

import (
	"github.com/kkevinchou/izzet/izzet/commandframe"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils"
	"github.com/kkevinchou/kitolib/metrics"
)

func (g *Game) GetSingleton() *singleton.Singleton {
	return g.singleton
}

func (g *Game) GetEntityByID(id int) entities.Entity {
	return g.entityManager.GetEntityByID(id)
}

func (g *Game) GetPlayer() *player.Player {
	if utils.IsServer() {
		panic("invalid call to GetPlayer() as server")
	}

	return g.GetPlayerByID(g.singleton.PlayerID)
}

func (g *Game) GetPlayerByID(id int) *player.Player {
	d := directory.GetDirectory().PlayerManager()
	player := d.GetPlayer(id)
	return player
}

func (g *Game) GetPlayerEntity() entities.Entity {
	if utils.IsServer() {
		panic("invalid call to GetPlayer() as server")
	}
	player := g.GetPlayer()
	return g.GetEntityByID(player.EntityID)
}

func (g *Game) GetPlayerEntityByID(id int) entities.Entity {
	player := g.GetPlayerByID(id)
	return g.GetEntityByID(player.EntityID)
}

func (g *Game) GetCamera() entities.Entity {
	return g.GetEntityByID(g.singleton.CameraID)
}

func (g *Game) RegisterEntities(entityList []entities.Entity) {
	for _, entity := range entityList {
		g.RegisterEntity(entity)
	}
}

func (g *Game) CommandFrame() int {
	return g.singleton.CommandFrame
}

func (g *Game) GetEventBroker() eventbroker.EventBroker {
	return g.eventBroker
}

func (g *Game) GetCommandFrameHistory() *commandframe.CommandFrameHistory {
	return g.commandFrameHistory
}

func (g *Game) MetricsRegistry() *metrics.MetricsRegistry {
	return g.metricsRegistry
}

func (g *Game) RegisterEntity(e entities.Entity) {
	g.entityManager.RegisterEntity(e)
}

func (g *Game) QueryEntity(componentFlags int) []entities.Entity {
	return g.entityManager.Query(componentFlags)
}

func (g *Game) UnregisterEntity(entity entities.Entity) {
	g.entityManager.UnregisterEntity(entity)
}

func (g *Game) UnregisterEntityByID(entityID int) {
	g.entityManager.UnregisterEntityByID(entityID)
}

func (g *Game) SetFocusedWindow(focusedWindow types.Window) {
	g.focusedWindow = focusedWindow
}

func (g *Game) GetFocusedWindow() types.Window {
	return g.focusedWindow
}

func (g *Game) GetWindowVisibility(window types.Window) bool {
	return g.windowVisibility[window]
}

func (g *Game) SetWindowVisibiilty(window types.Window, visible bool) {
	g.windowVisibility[window] = visible
}

func (g *Game) ToggleWindowVisibility(window types.Window) bool {
	g.windowVisibility[window] = !g.windowVisibility[window]
	return g.windowVisibility[window]
}

func (g *Game) SpatialPartition() *spatialpartition.SpatialPartition {
	return g.spatialPartition
}

func (g *Game) SetServerStats(serverStats map[string]string) {
	g.serverStats = serverStats
}

func (g *Game) ServerStats() map[string]string {
	return g.serverStats
}
