package components

import "github.com/go-gl/mathgl/mgl64"

type ThirdPersonControllerComponent struct {
	CameraID int

	Controlled bool
	Grounded   bool

	// Velocity      mgl64.Vec3
	BaseVelocity  mgl64.Vec3
	MovementSpeed float64

	ControllerVelocity mgl64.Vec3
	ZipVelocity        mgl64.Vec3
}

func (c *ThirdPersonControllerComponent) AddToComponentContainer(container *ComponentContainer) {
	container.ThirdPersonControllerComponent = c
}

func (c *ThirdPersonControllerComponent) ComponentFlag() int {
	return ComponentFlagThirdPersonController
}

func (c *ThirdPersonControllerComponent) Synchronized() bool {
	return false
}

func (c *ThirdPersonControllerComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *ThirdPersonControllerComponent) Serialize() []byte {
	panic("wat")
}
