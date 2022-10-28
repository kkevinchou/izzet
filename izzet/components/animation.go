package components

import "github.com/kkevinchou/izzet/lib/animation"

type AnimationComponent struct {
	Player *animation.AnimationPlayer
}

func (c *AnimationComponent) GetAnimationComponent() *AnimationComponent {
	return c
}

func (c *AnimationComponent) AddToComponentContainer(container *ComponentContainer) {
	container.AnimationComponent = c
}

func (c *AnimationComponent) ComponentFlag() int {
	return ComponentFlagAnimation
}

func (c *AnimationComponent) Synchronized() bool {
	return false
}

func (c *AnimationComponent) Load(bytes []byte) {
	panic("wat")
}

func (c *AnimationComponent) Serialize() []byte {
	panic("wat")
}
