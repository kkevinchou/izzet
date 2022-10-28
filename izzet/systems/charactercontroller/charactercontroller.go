package charactercontroller

import (
	"fmt"
	"time"

	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/netsync"
	"github.com/kkevinchou/izzet/izzet/singleton"
	"github.com/kkevinchou/izzet/izzet/systems/base"
	"github.com/kkevinchou/izzet/izzet/utils"
)

type World interface {
	GetSingleton() *singleton.Singleton
	GetEntityByID(id int) entities.Entity
	GetPlayerEntity() entities.Entity
	GetPlayer() *player.Player
}

type CharacterControllerSystem struct {
	*base.BaseSystem
	world World
}

func NewCharacterControllerSystem(world World) *CharacterControllerSystem {
	return &CharacterControllerSystem{
		BaseSystem: &base.BaseSystem{},
		world:      world,
	}
}

func (s *CharacterControllerSystem) Update(delta time.Duration) {
	d := directory.GetDirectory()
	playerManager := d.PlayerManager()
	singleton := s.world.GetSingleton()

	var players []*player.Player
	if utils.IsClient() {
		players = []*player.Player{s.world.GetPlayer()}
	} else {
		players = playerManager.GetPlayers()
	}

	for _, player := range players {
		entity := s.world.GetEntityByID(player.EntityID)
		if entity == nil {
			fmt.Printf("character controller could not find player entity with id %d\n", player.EntityID)
			continue
		}

		cameraID := entity.GetComponentContainer().ThirdPersonControllerComponent.CameraID
		camera := s.world.GetEntityByID(cameraID)
		if camera == nil {
			fmt.Printf("character controller could not find camera with entity id %d\n", cameraID)
			continue
		}

		netsync.UpdateCharacterController(delta, entity, camera, singleton.PlayerInput[player.ID])
	}
}

func (s *CharacterControllerSystem) Name() string {
	return "CharacterControllerSystem"
}
