package types

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
)

const (
	JumpImpulse       = "jump"
	ControllerImpulse = "controller"
)

type Impulse struct {
	Vector      mgl64.Vec3
	ElapsedTime time.Duration

	// the decay fraction per second for an impulse
	DecayRate float64
}
