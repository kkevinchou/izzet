package components

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/behavior"
)

type AIState string

const (
	AIStateWalk   AIState = "WALK"
	AIStateIdle   AIState = "IDLE"
	AIStateAttack AIState = "ATTACK"
)

type AIComponent struct {
	// behaviorTree behavior.BehaviorTree
	LastUpdate  time.Time
	MovementDir mgl64.Quat
	// Velocity    mgl64.Vec3

	AIState AIState
}

func NewAIComponent(behaviorTree behavior.BehaviorTree) *AIComponent {
	return &AIComponent{
		LastUpdate:  time.Now(),
		MovementDir: mgl64.QuatRotate(0, mgl64.Vec3{0, 1, 0}),
		AIState:     AIStateIdle,
		// behaviorTree: behaviorTree,
	}
}

func (c *AIComponent) AddToComponentContainer(container *ComponentContainer) {
	container.AIComponent = c
}

func (c *AIComponent) ComponentFlag() int {
	return ComponentFlagAI
}

func (c *AIComponent) Synchronized() bool {
	return false
}

func (c *AIComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *AIComponent) Serialize() []byte {
	panic("wat")
}
