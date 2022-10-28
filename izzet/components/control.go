package components

// mostly just holds the id of the player that controls this thing
type ControlComponent struct {
	PlayerID int
}

func (c *ControlComponent) AddToComponentContainer(container *ComponentContainer) {
	container.ControlComponent = c
}

func (c *ControlComponent) ComponentFlag() int {
	return ComponentFlagControl
}

func (c *ControlComponent) Synchronized() bool {
	return false
}

func (c *ControlComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *ControlComponent) Serialize() []byte {
	panic("wat")
}
