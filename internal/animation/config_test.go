package animation

import (
	"strings"
	"testing"
)

type TestContext struct {
}

func testConditionParser[T any](name string) Condition[T] {
	return NewGameCondition(name, func(ctx T) bool {
		return true
	})
}

func TestNewAnimationStateMachineFromYAMLReader(t *testing.T) {
	config := `
initial: idle
states:
  idle:
    clip: Idle_Loop
    playRate: 1
`

	sm := NewAnimationStateMachine[TestContext](strings.NewReader(config), testConditionParser[TestContext])

	assertCurrentState(t, sm, "idle", "Idle_Loop", 1)
}

func TestNewAnimationStateMachineParsesBuiltInClipCompletedCondition(t *testing.T) {
	config := `
initial: idle
states:
  idle:
    clip: Idle_Loop
    playRate: 1
    transitions:
      - to: done
        when: [clipCompleted]
  done:
    clip: Done
    playRate: 1
`

	sm := NewAnimationStateMachine[TestContext](strings.NewReader(config), nil)

	assertCurrentState(t, sm, "idle", "Idle_Loop", 1)
}
