package animation

import (
	"strings"
	"testing"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
)

func TestConfigureAnimationStateMachineFromDefaultYAML(t *testing.T) {
	sm := iztanimation.NewAnimationStateMachine[AnimationContext]()

	ConfigureAnimationStateMachine(sm)

	if got, want := sm.CurrentAnimationState(), "airborne"; got != want {
		t.Fatalf("current animation state = %q, want %q", got, want)
	}
}

func TestConfigureAnimationStateMachineFromYAMLReader(t *testing.T) {
	sm := iztanimation.NewAnimationStateMachine[AnimationContext]()
	config := `
initial: idle
states:
  idle:
    clip: Idle_Loop
    playRate: 1
`

	ConfigureAnimationStateMachineFromYAML(sm, strings.NewReader(config))

	if got, want := sm.CurrentAnimationState(), "idle"; got != want {
		t.Fatalf("current animation state = %q, want %q", got, want)
	}
}
