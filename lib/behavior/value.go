package behavior

import "time"

type Value struct {
	Value any
}

func (v *Value) Tick(input any, state AIState, delta time.Duration) (any, Status) {
	return v.Value, SUCCESS
}

func (v *Value) Reset() {}
