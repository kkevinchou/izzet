package system

import (
	"time"

	"github.com/kkevinchou/izzet/internal/utils"
	animationparser "github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/types"
)

type AnimationSystem struct {
	app App
}

func NewAnimationSystem(app App) *AnimationSystem {
	return &AnimationSystem{app: app}
}

func (s *AnimationSystem) Name() string {
	return "AnimationSystem"
}

func (s *AnimationSystem) Update(delta time.Duration, world GameWorld) {
	for _, e := range world.Entities() {
		if e.Animation == nil {
			continue
		}

		if s.app.IsClient() && s.app.AppMode() == types.AppModeEditor {
			c := e.Animation
			player := c.AnimationPlayer
			if e.Animation.SelectedAnimation != "" {
				if c.LoopAnimation {
					if player.CurrentAnimation() != c.SelectedAnimation || player.NormalizedClipProgress() >= 1 {
						player.PlayClip(c.SelectedAnimation)
					}
					player.Update(delta)
				} else {
					if player.CurrentAnimation() != c.SelectedAnimation {
						player.PlayClip(c.SelectedAnimation)
					}
					player.SetCurrentAnimationFrame(c.SelectedAnimation, c.SelectedKeyFrame)
				}
			}
		} else if e.Animation.Mode == entity.AnimationModeStateMachine {
			if (s.app.IsClient() && s.app.AppMode() == types.AppModePlay && s.app.GetPlayerEntity().GetID() == e.GetID()) || s.app.IsServer() {
				if e.Kinematic == nil {
					continue
				}

				var ctx animationparser.GameContext
				ctx.Grounded = e.Kinematic.Grounded
				ctx.JumpTriggered = e.Kinematic.Jump
				ctx.Moving = !utils.Vec3IsZero(e.Kinematic.MoveIntent)
				ctx.Dead = e.Deadge

				if e.AttackComponent != nil {
					ctx.Attacking = e.AttackComponent.Attacking
				}
				if e.AimDownSightsComponent != nil && e.AimDownSightsComponent.Active {
					ctx.AimDownSights = true
					ctx.AimDownSightsFire = e.AimDownSightsComponent.Fire
				}

				transition, transitioned := e.Animation.AnimationStateMachine.Update(delta, e.Animation.AnimationPlayer, ctx)

				// store animation transitions in preparation for replication
				if s.app.IsServer() && transitioned {
					e.Animation.AnimationTransitions = append(
						e.Animation.AnimationTransitions,
						entity.ServerSideAnimationTransition{
							AnimationTransition: transition,
							GlobalCommandFrame:  s.app.CommandFrame(),
						},
					)
				}
			} else {
				// trigger transitions sent over from the server
				if e.Animation.ReplicatedAnimationTransition != nil {
					e.Animation.AnimationStateMachine.TriggerTransition(
						e.Animation.AnimationPlayer,
						e.Animation.ReplicatedAnimationTransition.Source,
						e.Animation.ReplicatedAnimationTransition.Destination,
					)
				}
				e.Animation.AnimationPlayer.Update(delta)
			}
		}
	}
}
