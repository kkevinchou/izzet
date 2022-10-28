package components

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/model"
)

type MeshComponent struct {
	Scale       mgl64.Mat4
	Orientation mgl64.Mat4
	Model       *model.Model
}

func (c *MeshComponent) GetMeshComponent() *MeshComponent {
	return c
}

func (c *MeshComponent) AddToComponentContainer(container *ComponentContainer) {
	container.MeshComponent = c
}

func (c *MeshComponent) ComponentFlag() int {
	return ComponentFlagMesh
}

func (c *MeshComponent) Synchronized() bool {
	return false
}

func (c *MeshComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *MeshComponent) Serialize() []byte {
	panic("wat")
}
