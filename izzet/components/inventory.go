package components

import (
	"github.com/kkevinchou/izzet/izzet/components/protogen/inventory"
	"google.golang.org/protobuf/proto"
)

const (
	inventoryWidth  = 3
	inventoryHeight = 2
)

type Item struct {
	ID int
}

type InventoryComponent struct {
	Data   *inventory.Inventory
	Width  int
	Height int
}

func NewInventoryComponent() *InventoryComponent {
	inv := &inventory.Inventory{}
	for i := 0; i < inventoryWidth*inventoryHeight; i++ {
		inv.Items = append(inv.Items, &inventory.InventorySlot{Id: -1, Index: int64(i)})
	}
	return &InventoryComponent{Data: inv, Width: inventoryWidth, Height: inventoryHeight}
}

func (c *InventoryComponent) Swap(a int, b int) {
	item := c.Data.Items[b]
	id := item.Id
	count := item.Count

	c.Data.Items[b].Id = c.Data.Items[a].Id
	c.Data.Items[b].Count = c.Data.Items[a].Count
	c.Data.Items[a].Id = id
	c.Data.Items[a].Count = count
}

func (c *InventoryComponent) Add(item *Item) {
	found := false
	var nextAvail *int
	for index, invItem := range c.Data.Items {
		if invItem.Id == int64(item.ID) {
			invItem.Count += 1
			found = true
		} else if invItem.Id == -1 && nextAvail == nil {
			localIndex := index
			nextAvail = &localIndex
		}
	}
	if !found && nextAvail != nil {
		c.Data.Items[*nextAvail].Id = int64(item.ID)
		c.Data.Items[*nextAvail].Count = 1
	}
}

func (c *InventoryComponent) AddToComponentContainer(container *ComponentContainer) {
	container.InventoryComponent = c
}

func (c *InventoryComponent) ComponentFlag() int {
	return ComponentFlagInventory
}

func (c *InventoryComponent) Synchronized() bool {
	return true
}

func (c *InventoryComponent) Load(bytes []byte) {
	h := &inventory.Inventory{}
	err := proto.Unmarshal(bytes, h)
	if err != nil {
		panic(err)
	}
	c.Data = h
}

func (c *InventoryComponent) Serialize() []byte {
	bytes, err := proto.Marshal(c.Data)
	if err != nil {
		panic(err)
	}
	return bytes
}
