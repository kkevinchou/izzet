package behavior

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/behavior"
)

type Position struct {
	// TODO: write a test for this
	filler bool // empty structs share the same pointer address, this field prevents the node cache from accidentally caching
}

type Positionable interface {
	Position() mgl64.Vec3
}

func (p *Position) Tick(input any, state behavior.AIState, delta time.Duration) (any, behavior.Status) {
	if positionable, ok := input.(Positionable); ok {
		return positionable.Position(), behavior.SUCCESS
	}
	return nil, behavior.FAILURE
}

func (v *Position) Reset() {}
