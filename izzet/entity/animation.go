package entity

import (
	"strings"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/modelspec"
	animationparser "github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/assets/handle"
)

const (
	AnimationKeyIdle   = "IDLE"
	AnimationKeyAttack = "ATTACk"
	AnimationKeyRun    = "RUN"
)

type AnimationComponent struct {
	AnimationHandle handle.Animation
	RootJointID     int
	Animations      map[string]*modelspec.AnimationSpec `json:"-"`

	SelectedAnimation string
	SelectedKeyFrame  int
	LoopAnimation     bool

	AnimationStateMachine *iztanimation.AnimationStateMachine[animationparser.GameContext]
	AnimationPlayer       *iztanimation.AnimationPlayer `json:"-"`

	// --- Replication ---

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

func NewAnimationComponent(animationHandle handle.Animation, ml *assets.AssetManager) *AnimationComponent {
	animations, joints, rootJointID := ml.GetAnimations(animationHandle)
	animationPlayer := iztanimation.NewAnimationPlayer()
	animationPlayer.Initialize(animations, joints[rootJointID])

	var animationStateMachine *iztanimation.AnimationStateMachine[animationparser.GameContext]
	if strings.Contains(animationHandle.String(), "velociraptor") {
		animationStateMachine = animationparser.NewRaptorAnimationStateMachine()
	} else {
		animationStateMachine = animationparser.NewPlayerAnimationStateMachine()
	}

	return &AnimationComponent{
		RootJointID:     rootJointID,
		AnimationHandle: animationHandle,
		Animations:      animations,

		AnimationPlayer:       animationPlayer,
		AnimationStateMachine: animationStateMachine,
	}
}
