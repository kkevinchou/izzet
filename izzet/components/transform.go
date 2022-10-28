package components

import (
	"github.com/go-gl/mathgl/mgl64"
)

type TransformComponent struct {
	Position    mgl64.Vec3
	Orientation mgl64.Quat
}

func (c *TransformComponent) AddToComponentContainer(container *ComponentContainer) {
	container.TransformComponent = c
}

func (c *TransformComponent) ComponentFlag() int {
	return ComponentFlagTransform
}

func (c *TransformComponent) Synchronized() bool {
	return false
}

func (c *TransformComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *TransformComponent) Serialize() []byte {
	panic("wat")
}
