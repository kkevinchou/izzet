package ability

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/utils"
	"github.com/kkevinchou/izzet/izzet/utils/entityutils"
	"github.com/kkevinchou/izzet/lib/input"
)

type World interface {
	CommandFrame() int
	GetSingleton() *singleton.Singleton
	GetEntityByID(int) entities.Entity
	RegisterEntities([]entities.Entity)
}

type AbilitySystem struct {
	*base.BaseSystem
	world World

	// probably put this in a component
	cooldowns map[string]int64
}

func NewAbilitySystem(world World) *AbilitySystem {
	return &AbilitySystem{
		world:     world,
		cooldowns: map[string]int64{},
	}
}

func (s *AbilitySystem) Update(delta time.Duration) {
	singleton := s.world.GetSingleton()
	playerManager := directory.GetDirectory().PlayerManager()

	for _, player := range playerManager.GetPlayers() {
		playerInput := singleton.PlayerInput[player.ID]
		entity := s.world.GetEntityByID(player.EntityID)
		if entity == nil {
			fmt.Printf("ability system couldn't find player %d\n", player.ID)
			continue
		}

		if entity == nil {
			continue
		}

		cc := entity.GetComponentContainer()

		if key, ok := playerInput.KeyboardInput[input.KeyboardKeyQ]; ok && key.Event == input.KeyboardEventDown {
			cooldownLookup := fmt.Sprintf("%d_%s", player.ID, input.KeyboardKeyQ)
			if time.Now().UnixMilli()-s.cooldowns[cooldownLookup] < 500 {
				continue
			}
			s.cooldowns[cooldownLookup] = time.Now().UnixMilli()

			cc.NotepadComponent.LastAction = components.ActionCast

			if utils.IsServer() {
				projSpeed := 200
				cc := entity.GetComponentContainer()
				direction := cc.TransformComponent.Orientation.Rotate(mgl64.Vec3{0, 0, -1})
				position := cc.TransformComponent.Position.Add(mgl64.Vec3{0, 15, 0}).Add(direction.Mul(10))
				proj := entityutils.Spawn(types.EntityTypeProjectile, position, cc.TransformComponent.Orientation)
				projcc := proj.GetComponentContainer()
				projcc.PhysicsComponent.Velocity = direction.Mul(float64(projSpeed))
				s.world.RegisterEntities([]entities.Entity{proj})
			}
		}
	}
}

func (s *AbilitySystem) Name() string {
	return "AbilitySystem"
}
