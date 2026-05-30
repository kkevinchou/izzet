package animation

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"sort"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
	"gopkg.in/yaml.v3"
)

//go:embed default_state_machine.yaml
var defaultStateMachineConfig []byte

type animationStateMachineConfig struct {
	Initial string                    `yaml:"initial"`
	States  map[string]animationState `yaml:"states"`
}

type animationState struct {
	Clip        string                `yaml:"clip"`
	PlayRate    float64               `yaml:"playRate"`
	Transitions []animationTransition `yaml:"transitions"`
}

type animationTransition struct {
	Name string   `yaml:"name"`
	To   string   `yaml:"to"`
	When []string `yaml:"when"`
}

type AnimationContext struct {
	Grounded      bool
	Airborne      bool
	JumpTriggered bool
	Moving        bool
	Player        *iztanimation.AnimationPlayer
}

func ConfigureAnimationStateMachine(sm iztanimation.AnimationStateMachine[AnimationContext]) {
	ConfigureAnimationStateMachineFromYAML(sm, bytes.NewReader(defaultStateMachineConfig))
}

func ConfigureAnimationStateMachineFromYAML(sm iztanimation.AnimationStateMachine[AnimationContext], configReader io.Reader) {
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
			conditions := make([]iztanimation.Condition[AnimationContext], 0, len(transition.When))
			for _, conditionName := range transition.When {
				conditions = append(conditions, newCondition(conditionName))
			}
			sm.RegisterTransition(transitionName(source, transition, i), source, transition.To, conditions...)
		}
	}

	sm.SetCurrentState(config.Initial)
}

func transitionName(source string, transition animationTransition, index int) string {
	if transition.Name != "" {
		return transition.Name
	}

	return fmt.Sprintf("%s_to_%s_%d", source, transition.To, index)
}

func newCondition(name string) iztanimation.Condition[AnimationContext] {
	switch name {
	case "moving":
		return &MovingCondition{}
	case "notMoving":
		return &NotMovingCondition{}
	case "jumpTriggered":
		return &JumpTriggeredCondition{}
	case "clipCompleted":
		return &ClipCompletedCondition{}
	case "airborne":
		return &AirborneCondition{}
	case "grounded":
		return &GroundedCondition{}
	default:
		panic(fmt.Sprintf("unknown animation condition %q", name))
	}
}
