package components

type RenderComponent struct {
	IsVisible bool
}

func (c *RenderComponent) AddToComponentContainer(container *ComponentContainer) {
	container.RenderComponent = c
}

func (c *RenderComponent) ComponentFlag() int {
	return ComponentFlagRender
}

func (c *RenderComponent) Synchronized() bool {
	return false
}

func (c *RenderComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *RenderComponent) Serialize() []byte {
	panic("wat")
}
