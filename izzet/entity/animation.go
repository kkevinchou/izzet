package entity

import (
	iztanimation "github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/assets"
)

const (
	AnimationKeyIdle   = "IDLE"
	AnimationKeyAttack = "ATTACk"
	AnimationKeyRun    = "RUN"
)

type AnimationComponent struct {
	AnimationHandle assets.AnimationHandle
	RootJointID     int
	Animations      map[string]*modelspec.AnimationSpec `json:"-"`

	SelectedAnimation string
	SelectedKeyFrame  int
	LoopAnimation     bool

	AnimationStateMachineID animation.StateMachineID
	AnimationStateMachine   *iztanimation.AnimationStateMachine[animation.GameContext]
	AnimationPlayer         *iztanimation.AnimationPlayer `json:"-"`

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

func NewAnimationComponent(am *assets.AssetManager, handle assets.AnimationHandle, id animation.StateMachineID) *AnimationComponent {
	c := &AnimationComponent{}
	InitializeAnimationComponent(c, am, handle, id, "")
	return c
}

func InitializeAnimationComponent(c *AnimationComponent, am *assets.AssetManager, handle assets.AnimationHandle, id animation.StateMachineID, startState string) {
	animations, joints, rootJointID := am.GetAnimations(handle)

	c.RootJointID = rootJointID
	c.AnimationHandle = handle
	c.Animations = animations
	c.AnimationStateMachineID = id

	player := iztanimation.NewAnimationPlayer()
	player.Initialize(animations, joints[rootJointID])

	stateMachine := animation.NewStateMachine(id)
	if startState != "" {
		stateMachine.SetCurrentState(startState)
	}
	stateMachine.SynchronizePlayer(player)

	c.AnimationPlayer = player
	c.AnimationStateMachine = stateMachine
}
