package components

// Mostly work as a type flag atm
type NetworkComponent struct {
}

func (c *NetworkComponent) AddToComponentContainer(container *ComponentContainer) {
	container.NetworkComponent = c
}

func (c *NetworkComponent) ComponentFlag() int {
	return ComponentFlagNetwork
}

func (c *NetworkComponent) Synchronized() bool {
	return false
}

func (c *NetworkComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *NetworkComponent) Serialize() []byte {
	panic("wat")
}
