package entity

import (
	iztanimation "github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/modelspec"
	animationparser "github.com/kkevinchou/izzet/izzet/animation"
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

	AnimationStateMachine *iztanimation.AnimationStateMachineImpl[animationparser.AnimationContext]
	AnimationPlayer       *iztanimation.AnimationPlayer `json:"-"`
}

func NewAnimationComponent(animationHandle string, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)
	animationPlayer := iztanimation.NewAnimationPlayer()
	animationPlayer.Initialize(animations, joints[rootJointID])
	animationStateMachine := iztanimation.NewAnimationStateMachine[animationparser.AnimationContext]()
	animationparser.ConfigureAnimationStateMachine(animationStateMachine)

	return &AnimationComponent{
		RootJointID:     rootJointID,
		AnimationHandle: animationHandle,
		Animations:      animations,
		AnimationNames:  make(map[string]string),

		AnimationPlayer:       animationPlayer,
		AnimationStateMachine: animationStateMachine,
	}
}
