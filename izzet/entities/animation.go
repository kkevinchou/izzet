package entities

import (
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/kitolib/animation"
)

type AnimationComponent struct {
	AnimationHandle string
	AnimationPlayer *animation.AnimationPlayer `json:"-"`
	RootJointID     int
}

func NewAnimationComponent(animationHandle string, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)
	animationPlayer := animation.NewAnimationPlayer()
	animationPlayer.Initialize(animations, joints[rootJointID])
	return &AnimationComponent{RootJointID: rootJointID, AnimationHandle: animationHandle, AnimationPlayer: animationPlayer}
}
