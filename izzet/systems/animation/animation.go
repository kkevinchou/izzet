package animation

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils"
	"github.com/kkevinchou/kitolib/libutils"
)

type World interface {
	QueryEntity(componentFlags int) []entities.Entity
	GetPlayerEntity() entities.Entity
	GetEntityByID(id int) entities.Entity
}

type AnimationSystem struct {
	*base.BaseSystem
	world World
}

func NewAnimationSystem(world World) *AnimationSystem {
	return &AnimationSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
	}
}

func (s *AnimationSystem) Update(delta time.Duration) {
	if utils.IsClient() {
		// play animations for the player
		playerEntity := s.world.GetPlayerEntity()
		cc := playerEntity.GetComponentContainer()
		if cc.AnimationComponent != nil {
			findAndPlayAnimation(delta, playerEntity)
			cc.AnimationComponent.Player.Update(delta)
		}

		// update the animation player for all other entities, relying on animation state
		// synchronization from the server
		for _, entity := range s.world.QueryEntity(components.ComponentFlagAnimation) {
			if entity.GetID() == playerEntity.GetID() {
				continue
			}
			entity.GetComponentContainer().AnimationComponent.Player.Update(delta)
		}
	} else {
		for _, entity := range s.world.QueryEntity(components.ComponentFlagAnimation) {
			findAndPlayAnimation(delta, entity)
			entity.GetComponentContainer().AnimationComponent.Player.Update(delta)
		}
	}
}

// findAndPlayAnimation takes an entity and finds the appropriate animation based on its state, then plays it
func findAndPlayAnimation(delta time.Duration, entity entities.Entity) {
	componentContainer := entity.GetComponentContainer()
	animationComponent := componentContainer.AnimationComponent
	player := animationComponent.Player

	if entity.Type() == types.EntityTypeBob {
		tpcComponent := componentContainer.ThirdPersonControllerComponent
		movementComponent := componentContainer.MovementComponent
		notepad := componentContainer.NotepadComponent

		var targetAnimation string
		if !libutils.Vec3IsZero(movementComponent.Velocity) {
			if tpcComponent.Grounded {
				if notepad.LastAction == components.ActionCast {
					player.PlayOnce("Cast1", "Walk", 250*time.Millisecond)
				} else {
					targetAnimation = "Walk"
					player.PlayAndBlendAnimation(targetAnimation, 250*time.Millisecond)
				}
			} else {
				targetAnimation = "Falling"
				player.PlayAndBlendAnimation(targetAnimation, 250*time.Millisecond)
			}
		} else {
			targetAnimation = "Idle"
			if notepad.LastAction == components.ActionCast {
				targetAnimation = "Cast1"
				player.PlayOnce(targetAnimation, "Idle", 250*time.Millisecond)
			} else {
				player.PlayAndBlendAnimation(targetAnimation, 250*time.Millisecond)
			}
		}
	} else if entity.Type() == types.EntityTypeEnemy {
		aiComponent := entity.GetComponentContainer().AIComponent
		if aiComponent.AIState == components.AIStateIdle {
			player.PlayAndBlendAnimation("Idle", 250*time.Millisecond)
		} else if aiComponent.AIState == components.AIStateWalk {
			player.PlayAndBlendAnimation("Walk", 250*time.Millisecond)
		} else if aiComponent.AIState == components.AIStateAttack {
			player.PlayAndBlendAnimation("Punch", 250*time.Millisecond)
		} else {
			player.PlayAndBlendAnimation("Idle", 250*time.Millisecond)
		}
	}
}

func (s *AnimationSystem) Name() string {
	return "AnimationSystem"
}
