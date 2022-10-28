package components

type CameraComponent struct {
	FollowTargetEntityID int
	FollowDistance       float64
	MaxFollowDistance    float64
	YOffset              float64

	// this zoom stuff probably doesn't belong here
	ZoomSpeed float64
}

func (c *CameraComponent) AddToComponentContainer(container *ComponentContainer) {
	container.CameraComponent = c
}

func (c *CameraComponent) ComponentFlag() int {
	return ComponentFlagCamera
}

func (c *CameraComponent) Synchronized() bool {
	return false
}

func (c *CameraComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *CameraComponent) Serialize() []byte {
	panic("wat")
}
