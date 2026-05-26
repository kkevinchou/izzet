package animationv2

import (
	"time"
)

type App interface {
}

type World interface {
}

// Condition

type Condition interface {
	Evaluate(app App, world World, ctx AnimationContext) bool
	Name() string
}

type MovingCondition struct {
}

func (c *MovingCondition) Name() string {
	return "MovingCondition"
}

func (c *MovingCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.Moving
}

type NotMovingCondition struct {
}

func (c *NotMovingCondition) Name() string {
	return "NotMovingCondition"
}

func (c *NotMovingCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return !ctx.Moving
}

type JumpTriggeredCondition struct {
}

func (c *JumpTriggeredCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.JumpTriggered
}

func (c *JumpTriggeredCondition) Name() string {
	return "JumpTriggeredCondition"
}

type ClipCompletedCondition struct {
}

func (c *ClipCompletedCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.Player.elapsedTime >= ctx.Player.currentAnimation.Length
}

func (c *ClipCompletedCondition) Name() string {
	return "ClipCompletedCondition"
}

type AirborneCondition struct {
}

func (c *AirborneCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.Airborne
}

func (c *AirborneCondition) Name() string {
	return "AirborneCondition"
}

type GroundedCondition struct {
}

func (c *GroundedCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	return ctx.Grounded
}

func (c *GroundedCondition) Name() string {
	return "GroundedCondition"
}

// Animation State

type AnimationState struct {
	Name     string
	ClipName string
	PlayRate float64
}

// Transition

type Transition struct {
	name       string
	source     *AnimationState
	target     *AnimationState
	conditions []Condition
}

func NewTransition(name string, source, target *AnimationState) *Transition {
	return &Transition{name: name, source: source, target: target}
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

	Grounded      bool
	Airborne      bool
	JumpTriggered bool
	Moving        bool
}

type AnimationStateMachine struct {
	currentState *AnimationState
	transitions  []*Transition
}

func NewAnimationStateMachine() *AnimationStateMachine {
	idle := &AnimationState{Name: "idle", ClipName: "Idle_Loop", PlayRate: 1}
	airborne := &AnimationState{Name: "airborne", ClipName: "Jump_Loop", PlayRate: 1.5}
	jumpStart := &AnimationState{Name: "jumpStart", ClipName: "Jump_Start", PlayRate: 2}
	jumpLand := &AnimationState{Name: "jumpLand", ClipName: "Jump_Land", PlayRate: 1.5}
	sprint := &AnimationState{Name: "sprint", ClipName: "Sprint_Loop", PlayRate: 1}

	sm := &AnimationStateMachine{}
	sm.currentState = idle

	idleIdleTransition := NewTransition("idleIdleTransition", idle, idle)
	idleIdleTransition.AddCondition(&ClipCompletedCondition{})
	idleIdleTransition.AddCondition(&GroundedCondition{})

	idleJumpStartTransition := NewTransition("idleJumpStartTransition", idle, jumpStart)
	idleJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	sprintJumpStartTransition := NewTransition("sprintJumpStartTransition", sprint, jumpStart)
	sprintJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	idleSprintTransition := NewTransition("idleSprintTransition", idle, sprint)
	idleSprintTransition.AddCondition(&MovingCondition{})

	sprintSprintTransition := NewTransition("sprintSprintTransition", sprint, sprint)
	sprintSprintTransition.AddCondition(&MovingCondition{})
	sprintSprintTransition.AddCondition(&GroundedCondition{})
	sprintSprintTransition.AddCondition(&ClipCompletedCondition{})

	sprintIdleTransition := NewTransition("sprintIdleTransition", sprint, idle)
	sprintIdleTransition.AddCondition(&NotMovingCondition{})
	sprintIdleTransition.AddCondition(&GroundedCondition{})

	airborneAirborneTransition := NewTransition("airborneAirborneTransition", airborne, airborne)
	airborneAirborneTransition.AddCondition(&ClipCompletedCondition{})

	jumpStartAirborneTransition := NewTransition("jumpStartAirborneTransition", jumpStart, airborne)
	jumpStartAirborneTransition.AddCondition(&ClipCompletedCondition{})
	jumpStartAirborneTransition.AddCondition(&AirborneCondition{})

	airborneJumpLandTransition := NewTransition("airborneJumpLandTransition", airborne, jumpLand)
	airborneJumpLandTransition.AddCondition(&GroundedCondition{})

	jumpLandSprintTransition := NewTransition("jumpLandSprintTransition", jumpLand, sprint)
	jumpLandSprintTransition.AddCondition(&GroundedCondition{})
	jumpLandSprintTransition.AddCondition(&MovingCondition{})

	jumpLandIdleTransition := NewTransition("jumpLandIdleTransition", jumpLand, idle)
	jumpLandIdleTransition.AddCondition(&GroundedCondition{})
	jumpLandIdleTransition.AddCondition(&ClipCompletedCondition{})

	jumpLandJumpStartTransition := NewTransition("jumpLandJumpStartTransition", jumpLand, jumpStart)
	jumpLandJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	sm.transitions = append(sm.transitions, jumpLandJumpStartTransition)
	sm.transitions = append(sm.transitions, idleJumpStartTransition)
	sm.transitions = append(sm.transitions, sprintJumpStartTransition)
	sm.transitions = append(sm.transitions, idleSprintTransition)
	sm.transitions = append(sm.transitions, sprintSprintTransition)
	sm.transitions = append(sm.transitions, sprintIdleTransition)
	sm.transitions = append(sm.transitions, jumpStartAirborneTransition)
	sm.transitions = append(sm.transitions, airborneJumpLandTransition)
	sm.transitions = append(sm.transitions, jumpLandSprintTransition)
	sm.transitions = append(sm.transitions, jumpLandIdleTransition)
	sm.transitions = append(sm.transitions, idleIdleTransition)
	sm.transitions = append(sm.transitions, airborneAirborneTransition)

	return sm
}

func (sm *AnimationStateMachine) CurrentAnimationState() string {
	return sm.currentState.Name
}

func (sm *AnimationStateMachine) Update(delta time.Duration, app App, world World, ctx AnimationContext) {
	// TDOO - maybe find a better place to initialize the player
	if ctx.Player.CurrentAnimation() == "" {
		ctx.Player.SetPlayRate(sm.currentState.PlayRate)
		ctx.Player.PlayClip(sm.currentState.ClipName)
	}

	// i know the current state, i only need to look up the relevant transitions theoretically
	// for each transition
	// - determine when the current state's animation starts blending
	// - determine when to actually update the transition state
	// - these are all properties of transitions
	// - how do we handle multiple transitions happening? have a priority order?

	ctx.Player.Update(delta)
	for _, t := range sm.transitions {
		if sm.currentState.Name != t.source.Name {
			continue
		}
		if t.Evaluate(app, world, ctx) {
			sm.currentState = t.NextState()
			ctx.Player.SetPlayRate(sm.currentState.PlayRate)
			ctx.Player.PlayClip(sm.currentState.ClipName)
			break
		}
	}
}
