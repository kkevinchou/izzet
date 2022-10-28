package components

const (
	ComponentFlagAnimation             = 1 << 1
	ComponentFlagCamera                = 1 << 2
	ComponentFlagCollider              = 1 << 3
	ComponentFlagControl               = 1 << 4
	ComponentFlagFollow                = 1 << 5
	ComponentFlagMesh                  = 1 << 6
	ComponentFlagNetwork               = 1 << 7
	ComponentFlagPhysics               = 1 << 8
	ComponentFlagRender                = 1 << 9
	ComponentFlagThirdPersonController = 1 << 10
	ComponentFlagTransform             = 1 << 11
	ComponentFlagAI                    = 1 << 12
	ComponentFlagNotepad               = 1 << 13
	ComponentFlagHealth                = 1 << 14
	ComponentFlagLootDropper           = 1 << 15
	ComponentFlagLoot                  = 1 << 16
	ComponentFlagInventory             = 1 << 17
	ComponentFlagMovement              = 1 << 18
)

type Component interface {
	AddToComponentContainer(container *ComponentContainer)
	ComponentFlag() int
	Synchronized() bool
	Load(bytes []byte)
	Serialize() []byte
}

type ComponentContainer struct {
	bitflags     int
	componentMap map[int]Component

	AIComponent                    *AIComponent
	AnimationComponent             *AnimationComponent
	RenderComponent                *RenderComponent
	TransformComponent             *TransformComponent
	PhysicsComponent               *PhysicsComponent
	TopDownViewComponent           *TopDownViewComponent
	ThirdPersonControllerComponent *ThirdPersonControllerComponent
	CameraComponent                *CameraComponent
	NetworkComponent               *NetworkComponent
	MeshComponent                  *MeshComponent
	ColliderComponent              *ColliderComponent
	ControlComponent               *ControlComponent
	NotepadComponent               *NotepadComponent
	HealthComponent                *HealthComponent
	LootDropperComponent           *LootDropperComponent
	LootComponent                  *LootComponent
	InventoryComponent             *InventoryComponent
	MovementComponent              *MovementComponent
}

func NewComponentContainer(components ...Component) *ComponentContainer {
	container := &ComponentContainer{
		componentMap: map[int]Component{},
	}
	for _, component := range components {
		component.AddToComponentContainer(container)
		container.componentMap[component.ComponentFlag()] = component
		container.bitflags |= component.ComponentFlag()
	}
	return container
}

func (cc *ComponentContainer) Load(components map[int][]byte) {
	for id, bytes := range components {
		cc.componentMap[id].Load(bytes)
	}
}

func (cc *ComponentContainer) Serialize() map[int][]byte {
	results := map[int][]byte{}
	for id, c := range cc.componentMap {
		if c.Synchronized() {
			results[id] = c.Serialize()
		}
	}
	return results
}

func (cc *ComponentContainer) SetBitFlag(b int) {
	cc.bitflags |= b
}

func (cc *ComponentContainer) MatchBitFlags(b int) bool {
	return b&cc.bitflags == b
}
