package animation

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/kkevinchou/izzet/internal/iztlog"
	"gopkg.in/yaml.v3"
)

type AnimationTransition struct {
	Source      string
	Destination string
}

type AnimationState struct {
	Name     string
	ClipName string
	PlayRate float64
}

type AnimationStateMachine[T any] struct {
	CurrentState *AnimationState `json:"CurrentState"`

	states          map[string]*AnimationState
	transitionNames map[string]struct{}
	transitions     []transition[T]
}

func NewAnimationStateMachine[T any](configReader io.Reader, conditionParser func(string) Condition[T]) *AnimationStateMachine[T] {
	sm := &AnimationStateMachine[T]{
		states:          map[string]*AnimationState{},
		transitionNames: map[string]struct{}{},
	}

	var config animationStateMachineConfig
	if err := yaml.NewDecoder(configReader).Decode(&config); err != nil {
		panic(err)
	}

	stateNames := make([]string, 0, len(config.States))
	for stateName := range config.States {
		stateNames = append(stateNames, stateName)
	}
	sort.Strings(stateNames)

	for _, stateName := range stateNames {
		state := config.States[stateName]
		sm.RegisterAnimationState(stateName, state.Clip, state.PlayRate)
	}

	for _, source := range stateNames {
		state := config.States[source]
		for i, transition := range state.Transitions {
			conditions := make([]Condition[T], 0, len(transition.When))
			for _, conditionName := range transition.When {
				conditions = append(conditions, parseCondition(conditionName, conditionParser))
			}
			sm.RegisterTransition(transitionName(source, transition, i), source, transition.To, conditions...)
		}
	}

	sm.SetCurrentState(config.Initial)

	return sm
}

// SynchronizePlayer aligns a player with the state machine's current state
func (sm *AnimationStateMachine[T]) SynchronizePlayer(player *AnimationPlayer) {
	player.PlayClip(sm.CurrentState.ClipName)
}

func (sm *AnimationStateMachine[T]) RegisterAnimationState(name, clipName string, playRate float64) {
	if name == "" {
		panic("animation state name cannot be empty")
	}

	if _, ok := sm.states[name]; ok {
		panic(fmt.Sprintf("animation state %q is already registered", name))
	}

	sm.states[name] = &AnimationState{Name: name, ClipName: clipName, PlayRate: playRate}
}

func (sm *AnimationStateMachine[T]) RegisterTransition(name, sourceStateName, targetStateName string, conditions ...Condition[T]) {
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
		if condition.eval == nil {
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

func (sm *AnimationStateMachine[T]) SetCurrentState(name string) {
	state, ok := sm.states[name]
	if !ok {
		panic(fmt.Sprintf("unknown animation state %q", name))
	}

	sm.CurrentState = state
}

func (sm *AnimationStateMachine[T]) TriggerTransition(player *AnimationPlayer, source, destination string) {
	for _, t := range sm.transitions {
		if sm.CurrentState.Name != t.SourceState().Name {
			continue
		}

		if t.SourceState().Name == source && t.NextState().Name == destination {
			sm.CurrentState = t.NextState()
			player.SetPlayRate(sm.CurrentState.PlayRate)
			player.BlendClip(sm.CurrentState.ClipName, 100*time.Millisecond)
			return
		}
	}
	iztlog.ClientLogger.Info("failed to trigger transition, hard setting animation state", "current", sm.CurrentState.Name, "src", source, "dst", destination)
	sm.SetCurrentState(destination)
}

func (sm *AnimationStateMachine[T]) Update(delta time.Duration, player *AnimationPlayer, gameCtx T) (AnimationTransition, bool) {
	if sm.CurrentState == nil {
		return AnimationTransition{}, false
	}

	player.Update(delta)
	ctx := evalContext[T]{
		game:   gameCtx,
		player: player,
	}

	for _, t := range sm.transitions {
		if sm.CurrentState.Name != t.SourceState().Name {
			continue
		}

		if t.Evaluate(ctx) {
			var blend bool
			stateTransition := AnimationTransition{
				Source:      sm.CurrentState.Name,
				Destination: t.NextState().Name,
			}
			if sm.CurrentState.Name != t.NextState().Name {
				blend = true
			}
			sm.CurrentState = t.NextState()
			player.SetPlayRate(sm.CurrentState.PlayRate)

			if blend {
				player.BlendClip(sm.CurrentState.ClipName, 100*time.Millisecond)
			} else {
				player.PlayClip(sm.CurrentState.ClipName)
			}
			return stateTransition, true
		}
	}
	return AnimationTransition{}, false
}

func transitionName(source string, transition transitionConfig, index int) string {
	return fmt.Sprintf("%s_to_%s_%d", source, transition.To, index)
}

func parseCondition[T any](name string, conditionParser func(string) Condition[T]) Condition[T] {
	switch name {
	case ConditionClipCompleted:
		return ClipCompletedCondition[T]()
	}

	if conditionParser == nil {
		panic(fmt.Sprintf("unknown animation condition %q", name))
	}

	return conditionParser(name)
}
