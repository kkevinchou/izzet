package world

import (
	"slices"

	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entity"
)

func (g *GameWorld) AddEntity(e *entity.Entity) {
	if _, ok := g.entities[e.GetID()]; ok {
		return
	}
	g.entities[e.ID] = e
	g.addEntityToSortedList(e)
}

// we maintain a sorted entity list to provide deterministic entity iteration
// this prevents weird non deterministic bugs which can occur, especially in
// a multiplayer game with state synchronization
func (g *GameWorld) addEntityToSortedList(e *entity.Entity) {
	i, found := slices.BinarySearchFunc(g.sortedEntities, e, func(a, b *entity.Entity) int {
		return a.ID - b.ID
	})
	if !found {
		g.sortedEntities = slices.Insert(g.sortedEntities, i, e)
	}
}

func (g *GameWorld) DeleteEntity(entityID int) {
	e := g.GetEntityByID(entityID)
	if e == nil {
		return
	}

	for _, child := range e.Children {
		entity.RemoveParent(child)
		g.DeleteEntity(child.GetID())
	}

	entity.RemoveParent(e)
	g.spatialPartition.DeleteEntity(e.ID)
	delete(g.entities, e.ID)

	g.removeEntityFromSortedList(e.ID)
}

func (g *GameWorld) removeEntityFromSortedList(entityID int) {
	i, found := slices.BinarySearchFunc(g.sortedEntities, entityID, func(e *entity.Entity, id int) int {
		return e.ID - id
	})
	if found {
		g.sortedEntities = slices.Delete(g.sortedEntities, i, i+1)
	}
}

func (g *GameWorld) GetEntityByID(id int) *entity.Entity {
	return g.entities[id]
}

func (g *GameWorld) Entities() []*entity.Entity {
	return g.sortedEntities
}

func (g *GameWorld) CommandFrame() int {
	return g.commandFrameCount
}

func (g *GameWorld) IncrementCommandFrameCount() {
	g.commandFrameCount++
}

func (g *GameWorld) Lights() []*entity.Entity {
	allEntities := g.Entities()
	result := []*entity.Entity{}
	for _, e := range allEntities {
		if e.LightInfo != nil {
			result = append(result, e)
		}
	}
	return result
}

func (g *GameWorld) SpatialPartition() *spatialpartition.SpatialPartition {
	return g.spatialPartition
}
