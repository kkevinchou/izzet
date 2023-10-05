package entities

import "github.com/kkevinchou/kitolib/animation"

type AnimationComponent struct {
	AnimationHandle string
	AnimationPlayer *animation.AnimationPlayer
	RootJointID     int
}
