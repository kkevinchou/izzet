package world

import (
	"sort"

	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entity"
)

func (g *GameWorld) AddEntity(e *entity.Entity) {
	if _, ok := g.entities[e.GetID()]; ok {
		return
	}
	g.entities[e.ID] = e
}

func (g *GameWorld) DestroyEntity(entityID int) {
	if _, ok := g.entities[entityID]; !ok {
		return
	}
	delete(g.entities, entityID)
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
}

func (g *GameWorld) GetEntityByID(id int) *entity.Entity {
	return g.entities[id]
}

func (g *GameWorld) Entities() []*entity.Entity {
	if g.sortFrame != g.CommandFrame() {
		g.sortFrame = g.CommandFrame()

		var ids []int
		for id := range g.entities {
			ids = append(ids, id)
		}

		sort.Ints(ids)

		entities := []*entity.Entity{}
		for _, id := range ids {
			entities = append(entities, g.entities[id])
		}
		g.sortedEntities = entities
	}

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
