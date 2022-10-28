package combat

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils"
)

type World interface {
	QueryEntity(componentFlags int) []entities.Entity
	GetEntityByID(id int) entities.Entity
	CommandFrame() int
	UnregisterEntity(entity entities.Entity)
	GetEventBroker() eventbroker.EventBroker
}

type CombatSystem struct {
	*base.BaseSystem

	world World
}

func NewCombatSystem(world World) *CombatSystem {
	return &CombatSystem{
		world: world,
	}
}

func (s *CombatSystem) Update(delta time.Duration) {
	if utils.IsClient() {
		return
	}

	// handle fireball collisions
	for _, entity := range s.world.QueryEntity(components.ComponentFlagCollider) {
		if entity.Type() == types.EntityTypeProjectile {
			contacts := entity.GetComponentContainer().ColliderComponent.Contacts
			if len(contacts) == 0 {
				continue
			}

			for e2ID, _ := range contacts {
				e2 := s.world.GetEntityByID(e2ID)
				health := e2.GetComponentContainer().HealthComponent
				if health != nil {
					health.Data.Value -= 50
				}
			}

			event := &events.UnregisterEntityEvent{
				GlobalCommandFrame: s.world.CommandFrame(),
				EntityID:           entity.GetID(),
			}
			s.world.GetEventBroker().Broadcast(event)
		}
	}

	// handle death events
	for _, entity := range s.world.QueryEntity(components.ComponentFlagHealth) {
		if entity.GetComponentContainer().HealthComponent.Data.Value <= 0 {
			event := &events.UnregisterEntityEvent{
				GlobalCommandFrame: s.world.CommandFrame(),
				EntityID:           entity.GetID(),
			}
			s.world.GetEventBroker().Broadcast(event)
		}
	}
}

func (s *CombatSystem) Name() string {
	return "CombatSystem"
}
