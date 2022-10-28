package components

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/types"
)

type PhysicsComponent struct {
	Static        bool
	Velocity      mgl64.Vec3
	Grounded      bool
	IgnoreGravity bool

	// impulses have a name that can be reset or overwritten
	Impulses map[string]types.Impulse
}

func (c *PhysicsComponent) ApplyImpulse(name string, impulse types.Impulse) {
	c.Impulses[name] = impulse
}

func (c *PhysicsComponent) AddToComponentContainer(container *ComponentContainer) {
	container.PhysicsComponent = c
}

func (c *PhysicsComponent) ComponentFlag() int {
	return ComponentFlagPhysics
}

func (c *PhysicsComponent) Synchronized() bool {
	return false
}

func (c *PhysicsComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *PhysicsComponent) Serialize() []byte {
	panic("wat")
}
