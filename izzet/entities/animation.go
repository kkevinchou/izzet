package entities

import (
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/modelspec"
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
	Animations      map[string]*modelspec.AnimationSpec

	AnimationNames map[string]string
}

func NewAnimationComponent(animationHandle string, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)
	animationPlayer := animation.NewAnimationPlayer()
	animationPlayer.Initialize(animations, joints[rootJointID])

	return &AnimationComponent{
		RootJointID:     rootJointID,
		AnimationHandle: animationHandle,
		AnimationPlayer: animationPlayer,
		Animations:      animations,
		AnimationNames:  make(map[string]string),
	}
}
