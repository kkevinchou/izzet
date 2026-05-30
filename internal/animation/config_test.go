package animation

import (
	"strings"
	"testing"
)

type TestContext struct {
}

type testCondition[T any] struct {
	name string
}

func (t *testCondition[T]) Evaluate(ctx T) bool {
	return true
}

func (t *testCondition[T]) Name() string {
	return t.name
}

func testConditionParser[T any](name string) Condition[T] {
	return &testCondition[T]{name: name}
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
