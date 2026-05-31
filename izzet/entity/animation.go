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

	SelectedAnimation string
	SelectedKeyFrame  int
	LoopAnimation     bool

	AnimationStateMachine *iztanimation.AnimationStateMachine[animationparser.GameContext] `json:"-"`
	AnimationPlayer       *iztanimation.AnimationPlayer                                    `json:"-"`
}

func NewAnimationComponent(animationHandle string, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)
	animationPlayer := iztanimation.NewAnimationPlayer()
	animationPlayer.Initialize(animations, joints[rootJointID])
	animationStateMachine := animationparser.NewPlayerAnimationStateMachine()

	return &AnimationComponent{
		RootJointID:     rootJointID,
		AnimationHandle: animationHandle,
		Animations:      animations,
		AnimationNames:  make(map[string]string),

		AnimationPlayer:       animationPlayer,
		AnimationStateMachine: animationStateMachine,
	}
}
