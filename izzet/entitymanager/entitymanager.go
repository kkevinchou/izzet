package entitymanager

import (
	"github.com/kkevinchou/izzet/izzet/entities"
)

type EntityManager struct {
	entities  map[int]entities.Entity
	entityIDs []int
}

func NewEntityManager() *EntityManager {
	return &EntityManager{
		entities: map[int]entities.Entity{},
	}
}

func (em *EntityManager) RegisterEntity(e entities.Entity) {
	em.entities[e.GetID()] = e
	em.entityIDs = append(em.entityIDs, e.GetID())
}

func (em *EntityManager) GetEntityByID(id int) entities.Entity {
	return em.entities[id]
}

// TODO: cache queries
func (em *EntityManager) Query(componentFlags int) []entities.Entity {
	var matches []entities.Entity
	for _, id := range em.entityIDs {
		e := em.entities[id]
		cc := e.GetComponentContainer()
		if cc.MatchBitFlags(componentFlags) {
			matches = append(matches, e)
		}
	}

	return matches
}

func (em *EntityManager) UnregisterEntity(e entities.Entity) {
	em.UnregisterEntityByID(e.GetID())
}

func (em *EntityManager) UnregisterEntityByID(entityID int) {
	delete(em.entities, entityID)
	var newEntityIDs []int
	for _, id := range em.entityIDs {
		if id == entityID {
			continue
		}
		newEntityIDs = append(newEntityIDs, id)
	}
	em.entityIDs = newEntityIDs
}
