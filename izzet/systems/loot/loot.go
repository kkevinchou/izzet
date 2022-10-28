package loot

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/managers/eventbroker"
	"github.com/kkevinchou/izzet/izzet/mechanics/items"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils/entityutils"
)

type World interface {
	QueryEntity(componentFlags int) []entities.Entity
	RegisterEntities([]entities.Entity)
	GetEntityByID(id int) entities.Entity
	CommandFrame() int
	GetEventBroker() eventbroker.EventBroker
}

type LootSystem struct {
	*base.BaseSystem
	world   World
	modPool *items.ModPool
}

func NewLootSystem(world World) *LootSystem {
	modPool := items.NewModPool()

	for i := 0; i < 100; i++ {
		modPool.AddMod(&items.Mod{ID: i, AffixType: items.AffixTypePrefix})
	}
	for i := 1000; i < 1100; i++ {
		modPool.AddMod(&items.Mod{ID: i, AffixType: items.AffixTypeSuffix})
	}

	return &LootSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
		modPool:    modPool,
	}
}

func (s *LootSystem) Update(delta time.Duration) {
	lootEntities := s.world.QueryEntity(components.ComponentFlagLootDropper)

	// drop loot from entities who have reached 0 health
	for _, entity := range lootEntities {
		cc := entity.GetComponentContainer()
		ldComponent := cc.LootDropperComponent
		healthComponent := cc.HealthComponent
		if ldComponent == nil || healthComponent == nil {
			continue
		}

		if healthComponent.Data.Value > 0 {
			continue
		}

		rarity := items.SelectRarity(ldComponent.Rarities, ldComponent.RarityWeights)
		mods := s.modPool.ChooseMods(rarity)
		_ = mods

		lootbox := entityutils.Spawn(types.EntityTypeLootbox, cc.TransformComponent.Position.Add(mgl64.Vec3{0, 25, 0}), cc.TransformComponent.Orientation)
		s.world.RegisterEntities([]entities.Entity{lootbox})
	}

	// add loot to a player's inventory
	inventoryEntities := s.world.QueryEntity(components.ComponentFlagInventory)
	for _, entity := range inventoryEntities {
		cc := entity.GetComponentContainer()
		if cc.ColliderComponent == nil || len(cc.ColliderComponent.Contacts) == 0 {
			continue
		}

		for e2ID := range cc.ColliderComponent.Contacts {
			cEntity := s.world.GetEntityByID(e2ID)
			if cEntity.Type() == types.EntityTypeLootbox {
				cc.InventoryComponent.Add(&components.Item{ID: 69})
				event := &events.UnregisterEntityEvent{
					GlobalCommandFrame: s.world.CommandFrame(),
					EntityID:           cEntity.GetID(),
				}
				s.world.GetEventBroker().Broadcast(event)
			}
		}
	}
}

func (s *LootSystem) Name() string {
	return "LootSystem"
}
