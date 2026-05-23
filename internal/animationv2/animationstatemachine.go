package animationv2

import (
	"time"

	"github.com/kkevinchou/izzet/internal/iztlog"
)

type App interface {
}

type World interface {
}

// Condition

type Condition interface {
	Evaluate(app App, world World, ctx AnimationContext) bool
}

type ClipCompletedCondition struct {
}

func (c *ClipCompletedCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.Player.elapsedTime >= ctx.Player.currentAnimation.Length
}

type AirborneCondition struct {
}

func (c *AirborneCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.Airborne
}

type GroundedCondition struct {
}

func (c *GroundedCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.Grounded
}

// Animation State

type AnimationState struct {
	name     string
	clipName string
}

// Transition

type Transition struct {
	source     *AnimationState
	target     *AnimationState
	conditions []Condition
}

func NewTransition(source, target *AnimationState) *Transition {
	return &Transition{source: source, target: target}
}

func (t *Transition) AddCondition(c Condition) {
	t.conditions = append(t.conditions, c)
}

func (t *Transition) NextState() *AnimationState {
	return t.target
}

func (t *Transition) Evaluate(app App, world World, ctx AnimationContext) bool {
	for _, c := range t.conditions {
		if !c.Evaluate(app, world, ctx) {
			return false
		}
	}
	return true
}

// === STATES ===
//
// TODO - break this down into multiple animations. i don't think my assets
// would work well with blend trees. could experiment with this later using
// normalized timestamps and scaling animations to have the animation period
// match
//
// grounded locomotion
//	- blend tree based on move amount
//		- movement amount == 0 		play idle clip
//		- movement amount == 0.5 	play jog clip
//		- movement amount == 1	 	play sprint clip
//
// jump start
//	- play jump start clip
//
// airborne
//	- play jump loop clip
//
// === TRANSITIONS ===
//
// grounded -> jumpstart
//	- source = grounded locomotion
//	- destination = jump start
//
//	- condition
//		- player input for jump was accepted
//	- blend
//		- start normalized ts = 0.8
//
// jumpstart -> airborne
//	- source = jump start
//	- destination = airborne
//
//	- condition
//		- source normalized ts >= 1.0
//

type AnimationContext struct {
	Player *AnimationPlayer

	Grounded bool
	Airborne bool
}

type AnimationStateMachine struct {
	currentState *AnimationState
	transitions  []*Transition
}

func NewAnimationStateMachine() *AnimationStateMachine {
	idle := &AnimationState{name: "idle", clipName: "Idle_Loop"}
	airborne := &AnimationState{name: "airborne", clipName: "Jump_Loop"}

	sm := &AnimationStateMachine{}
	sm.currentState = idle

	idleIdleTransition := NewTransition(idle, idle)
	idleIdleTransition.AddCondition(&ClipCompletedCondition{})
	idleIdleTransition.AddCondition(&GroundedCondition{})

	airborneAirborneTransition := NewTransition(airborne, airborne)
	airborneAirborneTransition.AddCondition(&ClipCompletedCondition{})

	idleairborneTransition := NewTransition(idle, airborne)
	idleairborneTransition.AddCondition(&AirborneCondition{})

	airborneIdleTransition := NewTransition(airborne, idle)
	airborneIdleTransition.AddCondition(&GroundedCondition{})

	sm.transitions = append(sm.transitions, idleairborneTransition)
	sm.transitions = append(sm.transitions, airborneIdleTransition)
	sm.transitions = append(sm.transitions, idleIdleTransition)
	sm.transitions = append(sm.transitions, airborneAirborneTransition)

	return sm
}

func (sm *AnimationStateMachine) Update(delta time.Duration, app App, world World, ctx AnimationContext) {
	// TDOO - maybe find a better place to initialize the player
	if ctx.Player.CurrentAnimation() == "" {
		ctx.Player.PlayClip(sm.currentState.clipName)
	}

	// i know the current state, i only need to look up the relevant transitions theoretically
	// for each transition
	// - determine when the current state's animation starts blending
	// - determine when to actually update the transition state
	// - these are all properties of transitions
	// - how do we handle multiple transitions happening? have a priority order?

	ctx.Player.Update(delta)
	iztlog.ClientLogger.Info("Update")
	for _, t := range sm.transitions {
		if sm.currentState.name != t.source.name {
			continue
		}
		if t.Evaluate(app, world, ctx) {
			sm.currentState = t.NextState()
			ctx.Player.PlayClip(sm.currentState.clipName)
			iztlog.ClientLogger.Info("PlayClip")
			break
		}
	}
}
