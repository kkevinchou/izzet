package entity

import (
	"github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/animationv2"
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
	AnimationPlayer *animation.AnimationPlayer `json:"-"`
	RootJointID     int
	Animations      map[string]*modelspec.AnimationSpec `json:"-"`

	AnimationNames map[string]string

	// Animation V2
	AnimationStateMachine *animationv2.AnimationStateMachine
	AnimationPlayerV2     *animationv2.AnimationPlayer `json:"-"`
}

func NewAnimationComponent(animationHandle string, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)

	animationPlayer := animation.NewAnimationPlayer()
	animationPlayerV2 := animationv2.NewAnimationPlayer()

	animationPlayer.Initialize(animations, joints[rootJointID])
	animationPlayerV2.Initialize(animations, joints[rootJointID])

	return &AnimationComponent{
		RootJointID:     rootJointID,
		AnimationHandle: animationHandle,
		AnimationPlayer: animationPlayer,
		Animations:      animations,
		AnimationNames:  make(map[string]string),

		AnimationPlayerV2:     animationPlayerV2,
		AnimationStateMachine: animationv2.NewAnimationStateMachine(),
	}
}
