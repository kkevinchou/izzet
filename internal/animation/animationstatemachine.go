package animation

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
	debug bool
}

func (c *ClipCompletedCondition) Evaluate(app App, world World, ctx AnimationContext) bool {
	result := ctx.Player.NormalizedClipProgress() >= 1
	return result
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
	sprintEnter := &AnimationState{Name: "sprintEnter", ClipName: "Sprint_Enter", PlayRate: 1.5}
	sprintExit := &AnimationState{Name: "sprintExit", ClipName: "Sprint_Exit", PlayRate: 2}

	sm := &AnimationStateMachine{}
	sm.currentState = airborne

	// idle

	idleIdleTransition := NewTransition("idleIdleTransition", idle, idle)
	idleIdleTransition.AddCondition(&ClipCompletedCondition{})
	idleIdleTransition.AddCondition(&GroundedCondition{})

	idleJumpStartTransition := NewTransition("idleJumpStartTransition", idle, jumpStart)
	idleJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	idleSprintEnterTransition := NewTransition("idleSprintTransition", idle, sprintEnter)
	idleSprintEnterTransition.AddCondition(&MovingCondition{})
	idleIdleTransition.AddCondition(&GroundedCondition{})

	// sprint  enter

	sprintEnterJumpStartTransition := NewTransition("sprintEnterJumpStartTransition", sprintEnter, jumpStart)
	sprintEnterJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	sprintEnterSprintExitTransition := NewTransition("sprintEnterSprintExitTransition", sprintEnter, sprintExit)
	sprintEnterSprintExitTransition.AddCondition(&NotMovingCondition{})
	sprintEnterSprintExitTransition.AddCondition(&GroundedCondition{})

	sprintEnterSprintTransition := NewTransition("sprintEnterSprintTransition", sprintEnter, sprint)
	sprintEnterSprintTransition.AddCondition(&MovingCondition{})
	sprintEnterSprintTransition.AddCondition(&ClipCompletedCondition{debug: true})

	// sprint

	sprintJumpStartTransition := NewTransition("sprintJumpStartTransition", sprint, jumpStart)
	sprintJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	sprintSprintExitTransition := NewTransition("sprintSprintExit", sprint, sprintExit)
	sprintSprintExitTransition.AddCondition(&NotMovingCondition{})

	sprintSprintTransition := NewTransition("sprintSprintTransition", sprint, sprint)
	sprintSprintTransition.AddCondition(&MovingCondition{})
	sprintSprintTransition.AddCondition(&GroundedCondition{})
	sprintSprintTransition.AddCondition(&ClipCompletedCondition{})

	sprintIdleTransition := NewTransition("sprintIdleTransition", sprint, idle)
	sprintIdleTransition.AddCondition(&NotMovingCondition{})
	sprintIdleTransition.AddCondition(&GroundedCondition{})

	// sprint exit

	sprintExitJumpStartTransition := NewTransition("sprintExitJumpStartTransition", sprintExit, jumpStart)
	sprintExitJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	sprintExitSprintEnterTransition := NewTransition("sprintExitSprintEnter", sprintExit, sprintEnter)
	sprintExitSprintEnterTransition.AddCondition(&MovingCondition{})
	sprintExitSprintEnterTransition.AddCondition(&GroundedCondition{})

	sprintExitIdleTransition := NewTransition("sprintExitIdle", sprintExit, idle)
	sprintExitIdleTransition.AddCondition(&NotMovingCondition{})
	sprintExitIdleTransition.AddCondition(&GroundedCondition{})
	sprintExitIdleTransition.AddCondition(&ClipCompletedCondition{})

	// jump land

	jumpLandSprintEnterTransition := NewTransition("jumpLandSprintTransition", jumpLand, sprintEnter)
	jumpLandSprintEnterTransition.AddCondition(&GroundedCondition{})
	jumpLandSprintEnterTransition.AddCondition(&MovingCondition{})

	jumpLandIdleTransition := NewTransition("jumpLandIdleTransition", jumpLand, idle)
	jumpLandIdleTransition.AddCondition(&GroundedCondition{})
	jumpLandIdleTransition.AddCondition(&ClipCompletedCondition{})

	jumpLandJumpStartTransition := NewTransition("jumpLandJumpStartTransition", jumpLand, jumpStart)
	jumpLandJumpStartTransition.AddCondition(&JumpTriggeredCondition{})

	// jump start

	jumpStartAirborneTransition := NewTransition("jumpStartAirborneTransition", jumpStart, airborne)
	jumpStartAirborneTransition.AddCondition(&ClipCompletedCondition{})
	jumpStartAirborneTransition.AddCondition(&AirborneCondition{})

	// airborne

	airborneAirborneTransition := NewTransition("airborneAirborneTransition", airborne, airborne)
	airborneAirborneTransition.AddCondition(&ClipCompletedCondition{})

	airborneJumpLandTransition := NewTransition("airborneJumpLandTransition", airborne, jumpLand)
	airborneJumpLandTransition.AddCondition(&GroundedCondition{})

	// add transitions

	sm.transitions = append(sm.transitions, jumpLandJumpStartTransition)
	sm.transitions = append(sm.transitions, idleJumpStartTransition)
	sm.transitions = append(sm.transitions, sprintJumpStartTransition)
	sm.transitions = append(sm.transitions, sprintEnterJumpStartTransition)
	sm.transitions = append(sm.transitions, sprintEnterSprintExitTransition)
	sm.transitions = append(sm.transitions, sprintExitJumpStartTransition)
	sm.transitions = append(sm.transitions, idleSprintEnterTransition)
	sm.transitions = append(sm.transitions, sprintExitSprintEnterTransition)
	sm.transitions = append(sm.transitions, sprintExitIdleTransition)
	sm.transitions = append(sm.transitions, sprintEnterSprintTransition)
	sm.transitions = append(sm.transitions, sprintSprintExitTransition)
	sm.transitions = append(sm.transitions, sprintSprintTransition)
	sm.transitions = append(sm.transitions, sprintIdleTransition)
	sm.transitions = append(sm.transitions, jumpStartAirborneTransition)
	sm.transitions = append(sm.transitions, airborneJumpLandTransition)
	sm.transitions = append(sm.transitions, jumpLandSprintEnterTransition)
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
		// t.blend config
		if t.Evaluate(app, world, ctx) {
			sm.currentState = t.NextState()
			ctx.Player.SetPlayRate(sm.currentState.PlayRate)
			ctx.Player.PlayClip(sm.currentState.ClipName)
			break
		}
	}
}
