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

	if got, want := sm.CurrentAnimationState(), "idle"; got != want {
		t.Fatalf("current animation state = %q, want %q", got, want)
	}
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

	if got, want := sm.CurrentAnimationState(), "idle"; got != want {
		t.Fatalf("current animation state = %q, want %q", got, want)
	}
}
