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

	assertCurrentState(t, sm, "idle", "Idle_Loop", 1)
}

func TestAnimationStateMachineRegistrationValidation(t *testing.T) {
	sm := &AnimationStateMachine[struct{}]{
		states:          map[string]*AnimationState{},
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

func assertCurrentState[T any](t *testing.T, sm *AnimationStateMachine[T], wantName, wantClip string, wantPlayRate float64) {
	t.Helper()

	state := sm.CurrentAnimationState()
	if state == nil {
		t.Fatal("current animation state is nil")
	}
	if state.Name != wantName {
		t.Fatalf("current animation state name = %q, want %q", state.Name, wantName)
	}
	if state.ClipName != wantClip {
		t.Fatalf("current animation state clip = %q, want %q", state.ClipName, wantClip)
	}
	if state.PlayRate != wantPlayRate {
		t.Fatalf("current animation state play rate = %v, want %v", state.PlayRate, wantPlayRate)
	}
}
