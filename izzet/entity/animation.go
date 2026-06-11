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

	// AnimationNames map[string]string

	SelectedAnimation string
	SelectedKeyFrame  int
	LoopAnimation     bool

	AnimationStateMachine *iztanimation.AnimationStateMachine[animationparser.GameContext]
	AnimationPlayer       *iztanimation.AnimationPlayer `json:"-"`

	// AnimationTransitions is the collection of animation transitions since the last
	// game state update was replicated to clients
	AnimationTransitions []ServerSideAnimationTransition

	// ReplicatedAnimationTransition is the animation transition we wish to apply this frame
	ReplicatedAnimationTransition *iztanimation.AnimationTransition
}

type ServerSideAnimationTransition struct {
	iztanimation.AnimationTransition
	GlobalCommandFrame int
}

func NewAnimationComponent(animationHandle string, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)
	animationPlayer := iztanimation.NewAnimationPlayer()
	animationPlayer.Initialize(animations, joints[rootJointID])
	animationStateMachine := animationparser.NewPlayerAnimationStateMachine()
	if animationHandle == string(EntityTypeVelociraptor) || animationHandle == "velociraptor" {
		animationStateMachine = animationparser.NewRaptorAnimationStateMachine()
	}

	return &AnimationComponent{
		RootJointID:     rootJointID,
		AnimationHandle: animationHandle,
		Animations:      animations,

		AnimationPlayer:       animationPlayer,
		AnimationStateMachine: animationStateMachine,
	}
}
