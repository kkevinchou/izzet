package components

import "github.com/go-gl/mathgl/mgl64"

type MovementComponent struct {
	Velocity mgl64.Vec3
}

func (c *MovementComponent) AddToComponentContainer(container *ComponentContainer) {
	container.MovementComponent = c
}

func (c *MovementComponent) ComponentFlag() int {
	return ComponentFlagMovement
}

func (c *MovementComponent) Synchronized() bool {
	return false
}

func (c *MovementComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *MovementComponent) Serialize() []byte {
	panic("wat")
}
