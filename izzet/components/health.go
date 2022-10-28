package components

import (
	"github.com/kkevinchou/izzet/izzet/components/protogen/health"
	"google.golang.org/protobuf/proto"
)

type HealthComponent struct {
	// Value  float64
	Data *health.Health
}

func NewHealthComponent(value float64) *HealthComponent {
	return &HealthComponent{
		Data: &health.Health{Value: value},
	}
}

func (c *HealthComponent) AddToComponentContainer(container *ComponentContainer) {
	container.HealthComponent = c
}

func (c *HealthComponent) ComponentFlag() int {
	return ComponentFlagHealth
}

func (c *HealthComponent) Synchronized() bool {
	return true
}

func (c *HealthComponent) Load(bytes []byte) {
	h := &health.Health{}
	err := proto.Unmarshal(bytes, h)
	if err != nil {
		panic(err)
	}
	c.Data = h
}

func (c *HealthComponent) Serialize() []byte {
	bytes, err := proto.Marshal(c.Data)
	if err != nil {
		panic(err)
	}
	return bytes
}
