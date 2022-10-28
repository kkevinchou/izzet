package components

import "github.com/kkevinchou/izzet/izzet/mechanics/items"

type LootDropperComponent struct {
	Rarities      []items.Rarity
	RarityWeights []int
}

func (c *LootDropperComponent) AddToComponentContainer(container *ComponentContainer) {
	container.LootDropperComponent = c
}

func (c *LootDropperComponent) ComponentFlag() int {
	return ComponentFlagLootDropper
}

func DefaultLootDropper() *LootDropperComponent {
	return &LootDropperComponent{
		Rarities:      []items.Rarity{items.RarityRare},
		RarityWeights: []int{1},
	}
}

func (c *LootDropperComponent) Synchronized() bool {
	return false
}

func (c *LootDropperComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *LootDropperComponent) Serialize() []byte {
	panic("wat")
}
