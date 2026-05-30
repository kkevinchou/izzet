package animation

import (
	"strings"
	"testing"
)

func TestAnimationStateMachineRegistration(t *testing.T) {
	config := `
initial: idle
states:
  idle:
    clip: Idle_Loop
    playRate: 1
    transitions:
      - to: run
  run:
    clip: Run_Loop
    playRate: 1.25
`

	sm := NewAnimationStateMachine[struct{}](strings.NewReader(config), nil)

	if got, want := sm.CurrentAnimationState(), "idle"; got != want {
		t.Fatalf("current animation state = %q, want %q", got, want)
	}
}

func TestAnimationStateMachineRegistrationValidation(t *testing.T) {
	sm := &AnimationStateMachine[struct{}]{
		states:          map[string]*animationState{},
		transitionNames: map[string]struct{}{},
	}

	sm.RegisterAnimationState("idle", "Idle_Loop", 1)
	assertPanics(t, func() { sm.RegisterAnimationState("idle", "Idle_Loop", 1) })
	assertPanics(t, func() { sm.RegisterTransition("idleRun", "idle", "run") })

	sm.RegisterAnimationState("run", "Run_Loop", 1)
	sm.RegisterTransition("idleRun", "idle", "run")

	assertPanics(t, func() { sm.RegisterTransition("idleRun", "idle", "run") })
	assertPanics(t, func() { sm.SetCurrentState("missing") })
}

func assertPanics(t *testing.T, f func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	f()
}
