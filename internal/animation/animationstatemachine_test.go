package animation

import "testing"

func TestAnimationStateMachineRegistration(t *testing.T) {
	sm := NewAnimationStateMachine[struct{}](nil, nil)

	sm.RegisterAnimationState("idle", "Idle_Loop", 1)
	sm.RegisterAnimationState("run", "Run_Loop", 1.25)
	sm.RegisterTransition("idleRun", "idle", "run")
	sm.SetCurrentState("idle")

	if got, want := sm.CurrentAnimationState(), "idle"; got != want {
		t.Fatalf("current animation state = %q, want %q", got, want)
	}
}

func TestAnimationStateMachineRegistrationValidation(t *testing.T) {
	sm := NewAnimationStateMachine[struct{}](nil, nil)

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
