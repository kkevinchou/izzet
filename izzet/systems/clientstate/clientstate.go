package clientstate

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/statebuffer"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils/entityutils"
	"github.com/kkevinchou/izzet/lib/metrics"
)

type World interface {
	CommandFrame() int
	GetSingleton() *singleton.Singleton
	GetEntityByID(int) entities.Entity
	RegisterEntities([]entities.Entity)
	GetPlayerEntity() entities.Entity
	GetPlayer() *player.Player
	MetricsRegistry() *metrics.MetricsRegistry
	UnregisterEntityByID(entityID int)
}

type ClientStateSystem struct {
	*base.BaseSystem
	world World
}

func NewClientStateSystem(world World) *ClientStateSystem {
	return &ClientStateSystem{
		world: world,
	}
}

func (s *ClientStateSystem) Update(delta time.Duration) {
	singleton := s.world.GetSingleton()

	state := singleton.StateBuffer.PeekEntityInterpolations(s.world.CommandFrame())
	if state != nil {
		spawnNewEntities(state, s.world)
	}

	state = singleton.StateBuffer.PullEntityInterpolations(s.world.CommandFrame())
	if state != nil {
		applyState(state, s.world)
	}
}

func spawnNewEntities(bufferedState *statebuffer.BufferedState, world World) {
	playerEntity := world.GetPlayerEntity()

	var newEntities []entities.Entity
	for _, snapshot := range bufferedState.InterpolatedEntities {
		if snapshot.ID == playerEntity.GetID() {
			continue
		}

		entity := world.GetEntityByID(snapshot.ID)
		if entity == nil {
			newEntity := entityutils.SpawnWithID(snapshot.ID, types.EntityType(snapshot.Type), snapshot.Position, snapshot.Orientation)
			newEntities = append(newEntities, newEntity)
		}
	}

	world.RegisterEntities(newEntities)
}

func applyState(bufferedState *statebuffer.BufferedState, world World) {
	playerEntity := world.GetPlayerEntity()
	for _, event := range bufferedState.Events {
		if event.Type == events.EventTypeUnregisterEntity {
			var e events.UnregisterEntityEvent
			knetwork.Deserialize(event.Bytes, &e)
			world.UnregisterEntityByID(e.EntityID)
		}
	}

	for _, entitySnapshot := range bufferedState.InterpolatedEntities {

		foundEntity := world.GetEntityByID(entitySnapshot.ID)
		if foundEntity == nil {
			fmt.Printf("[%d] failed to find entity with id %d type %d to interpolate\n", world.CommandFrame(), entitySnapshot.ID, entitySnapshot.Type)
		} else {
			cc := foundEntity.GetComponentContainer()
			cc.Load(entitySnapshot.Components)

			// do not synchronize transforms or animation
			if entitySnapshot.ID == playerEntity.GetID() {
				continue
			}

			cc.TransformComponent.Position = entitySnapshot.Position
			cc.TransformComponent.Orientation = entitySnapshot.Orientation
			if cc.MovementComponent != nil {
				cc.MovementComponent.Velocity = entitySnapshot.Velocity
			}
			if cc.AnimationComponent != nil {
				cc.AnimationComponent.Player.PlayAnimation(entitySnapshot.Animation)
			}
		}
	}
}

func (s *ClientStateSystem) Name() string {
	return "ClientStateSystem"
}
