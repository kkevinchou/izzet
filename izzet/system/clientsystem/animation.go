package clientsystem

import (
	"time"

	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/system"
	"github.com/kkevinchou/izzet/izzet/system/shared"
)

type EditorAnimationSystem struct {
}

func NewEditorAnimationSystem() *EditorAnimationSystem {
	return &EditorAnimationSystem{}
}

func (s *EditorAnimationSystem) Name() string {
	return "EditorAnimationSystem"
}

func (s *EditorAnimationSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, e := range world.Entities() {
		if e.Animation == nil {
			continue
		}

		updateEditorAnimationPreview(delta, e.Animation)
	}
}

type ClientAnimationSystem struct {
	app App
}

func NewClientAnimationSystem(app App) *ClientAnimationSystem {
	return &ClientAnimationSystem{app: app}
}

func (s *ClientAnimationSystem) Name() string {
	return "ClientAnimationSystem"
}

func (s *ClientAnimationSystem) Update(delta time.Duration, world system.GameWorld) {
	player := s.app.GetPlayerEntity()
	if player == nil {
		return
	}

	for _, e := range world.Entities() {
		if !shared.IsStateMachineAnimation(e) {
			continue
		}

		if e.GetID() == player.GetID() {
			shared.UpdateStateMachineAnimation(delta, e)
		} else {
			updateReplicatedAnimation(delta, e.Animation)
		}
	}
}

func updateEditorAnimationPreview(delta time.Duration, c *entity.AnimationComponent) {
	player := c.AnimationPlayer
	if c.SelectedAnimation == "" {
		return
	}

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

func updateReplicatedAnimation(delta time.Duration, c *entity.AnimationComponent) {
	if c.ReplicatedAnimationTransition != nil {
		c.AnimationStateMachine.TriggerTransition(
			c.AnimationPlayer,
			c.ReplicatedAnimationTransition.Source,
			c.ReplicatedAnimationTransition.Destination,
		)
	}
	c.AnimationPlayer.Update(delta)
}
