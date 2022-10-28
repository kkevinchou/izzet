package components

type LootComponent struct {
}

func (c *LootComponent) AddToComponentContainer(container *ComponentContainer) {
	container.LootComponent = c
}

func (c *LootComponent) ComponentFlag() int {
	return ComponentFlagLoot
}

func (c *LootComponent) Synchronized() bool {
	return false
}

func (c *LootComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *LootComponent) Serialize() []byte {
	panic("wat")
}
