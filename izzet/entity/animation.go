package entity

import (
	"github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/assets"
)

const (
	AnimationKeyIdle   = "IDLE"
	AnimationKeyAttack = "ATTACk"
	AnimationKeyRun    = "RUN"
)

type AnimationComponent struct {
	AnimationHandle string
	RootJointID     int
	Animations      map[string]*modelspec.AnimationSpec `json:"-"`

	AnimationNames map[string]string

	AnimationStateMachine *animation.AnimationStateMachine
	AnimationPlayer       *animation.AnimationPlayer `json:"-"`
}

func NewAnimationComponent(animationHandle string, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)
	animationPlayer := animation.NewAnimationPlayer()
	animationPlayer.Initialize(animations, joints[rootJointID])

	return &AnimationComponent{
		RootJointID:     rootJointID,
		AnimationHandle: animationHandle,
		Animations:      animations,
		AnimationNames:  make(map[string]string),

		AnimationPlayer:       animationPlayer,
		AnimationStateMachine: animation.NewAnimationStateMachine(),
	}
}
