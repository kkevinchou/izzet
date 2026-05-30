package animation

import (
	"fmt"
	"time"
)

type Condition[T any] interface {
	Evaluate(ctx T) bool
	Name() string
}

type animationState struct {
	Name     string
	ClipName string
	PlayRate float64
}

type transition[T any] interface {
	SourceState() *animationState
	NextState() *animationState
	Evaluate(ctx T) bool
}

type transitionImpl[T any] struct {
	name       string
	source     *animationState
	target     *animationState
	conditions []Condition[T]
}

func (t *transitionImpl[T]) AddCondition(c Condition[T]) {
	t.conditions = append(t.conditions, c)
}

func (t *transitionImpl[T]) SourceState() *animationState {
	return t.source
}

func (t *transitionImpl[T]) NextState() *animationState {
	return t.target
}

func (t *transitionImpl[T]) Evaluate(ctx T) bool {
	transition := true

	for _, c := range t.conditions {
		if !c.Evaluate(ctx) {
			transition = false
			break
		}
	}

	return transition
}

type AnimationStateMachine[T any] interface {
	RegisterAnimationState(name, clipName string, playRate float64)
	RegisterTransition(name, sourceStateName, targetStateName string, conditions ...Condition[T])
	SetCurrentState(name string)
}

type AnimationStateMachineImpl[T any] struct {
	currentState    *animationState
	states          map[string]*animationState
	transitionNames map[string]struct{}
	transitions     []transition[T]
}

func NewAnimationStateMachine[T any]() *AnimationStateMachineImpl[T] {
	return &AnimationStateMachineImpl[T]{
		states:          map[string]*animationState{},
		transitionNames: map[string]struct{}{},
	}
}

func (sm *AnimationStateMachineImpl[T]) RegisterAnimationState(name, clipName string, playRate float64) {
	if name == "" {
		panic("animation state name cannot be empty")
	}

	if _, ok := sm.states[name]; ok {
		panic(fmt.Sprintf("animation state %q is already registered", name))
	}

	sm.states[name] = &animationState{Name: name, ClipName: clipName, PlayRate: playRate}
}

func (sm *AnimationStateMachineImpl[T]) RegisterTransition(name, sourceStateName, targetStateName string, conditions ...Condition[T]) {
	if name == "" {
		panic("animation transition name cannot be empty")
	}

	if _, ok := sm.transitionNames[name]; ok {
		panic(fmt.Sprintf("animation transition %q is already registered", name))
	}

	source, ok := sm.states[sourceStateName]
	if !ok {
		panic(fmt.Sprintf("animation transition %q references unknown source state %q", name, sourceStateName))
	}

	target, ok := sm.states[targetStateName]
	if !ok {
		panic(fmt.Sprintf("animation transition %q references unknown target state %q", name, targetStateName))
	}

	for _, condition := range conditions {
		if condition == nil {
			panic(fmt.Sprintf("animation transition %q contains nil condition", name))
		}
	}

	t := &transitionImpl[T]{name: name, source: source, target: target}
	for _, condition := range conditions {
		t.AddCondition(condition)
	}

	sm.transitions = append(sm.transitions, t)
	sm.transitionNames[name] = struct{}{}
}

func (sm *AnimationStateMachineImpl[T]) SetCurrentState(name string) {
	state, ok := sm.states[name]
	if !ok {
		panic(fmt.Sprintf("unknown animation state %q", name))
	}

	sm.currentState = state
}

func (sm *AnimationStateMachineImpl[T]) CurrentAnimationState() string {
	if sm.currentState == nil {
		return ""
	}

	return sm.currentState.Name
}

func (sm *AnimationStateMachineImpl[T]) Update(delta time.Duration, player *AnimationPlayer, ctx T) {
	if sm.currentState == nil {
		return
	}

	// TDOO - maybe find a better place to initialize the player
	if player.CurrentAnimation() == "" {
		player.SetPlayRate(sm.currentState.PlayRate)
		player.PlayClip(sm.currentState.ClipName)
	}

	player.Update(delta)
	for _, t := range sm.transitions {
		if sm.currentState.Name != t.SourceState().Name {
			continue
		}

		if t.Evaluate(ctx) {
			var blend bool
			if sm.currentState != t.NextState() {
				blend = true
			}
			sm.currentState = t.NextState()
			player.SetPlayRate(sm.currentState.PlayRate)

			if blend {
				player.BlendClip(sm.currentState.ClipName, 100*time.Millisecond)
			} else {
				player.PlayClip(sm.currentState.ClipName)
			}
			break
		}
	}
}
