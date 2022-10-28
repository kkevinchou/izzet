package izzet

import (
	"fmt"
	"math/rand"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/directory"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems/ability"
	"github.com/kkevinchou/izzet/izzet/systems/ai"
	"github.com/kkevinchou/izzet/izzet/systems/animation"
	"github.com/kkevinchou/izzet/izzet/systems/bookkeeping"
	"github.com/kkevinchou/izzet/izzet/systems/charactercontroller"
	"github.com/kkevinchou/izzet/izzet/systems/collision"
	"github.com/kkevinchou/izzet/izzet/systems/combat"
	"github.com/kkevinchou/izzet/izzet/systems/loot"
	"github.com/kkevinchou/izzet/izzet/systems/networkdispatch"
	"github.com/kkevinchou/izzet/izzet/systems/networkupdate"
	"github.com/kkevinchou/izzet/izzet/systems/physics"
	"github.com/kkevinchou/izzet/izzet/systems/playerinput"
	"github.com/kkevinchou/izzet/izzet/systems/playerregistration"
	"github.com/kkevinchou/izzet/izzet/systems/preframe"
	"github.com/kkevinchou/izzet/izzet/systems/rpcreceiver"
	"github.com/kkevinchou/izzet/lib/assets"
)

func NewServerGame(assetsDirectory string) *Game {
	initSeed()
	settings.CurrentGameMode = settings.GameModeServer

	g := NewBaseGame()

	serverSystemSetup(g, assetsDirectory)
	initialEntities := serverEntitySetup(g)
	g.RegisterEntities(initialEntities)

	return g
}

func serverEntitySetup(g *Game) []entities.Entity {
	scene := entities.NewScene()

	enemies := []entities.Entity{}
	for i := 0; i < 5; i++ {
		enemy := entities.NewEnemy()
		x := rand.Intn(1000) - 500
		z := rand.Intn(1000) - 500
		enemy.GetComponentContainer().TransformComponent.Position = mgl64.Vec3{float64(x), 0, float64(z)}
		enemies = append(enemies, enemy)
	}

	lootbox := entities.NewLootbox()
	furniture := entities.NewStaticRigidBody()
	furniture.GetComponentContainer().TransformComponent.Position[0] = 200
	furniture.GetComponentContainer().TransformComponent.Position[2] = 150

	entities := []entities.Entity{
		scene,
		lootbox,
		furniture,
	}
	entities = append(entities, enemies...)
	return entities
}

func serverSystemSetup(g *Game, assetsDirectory string) {
	d := directory.GetDirectory()

	playerManager := player.NewPlayerManager(g)
	d.RegisterPlayerManager(playerManager)

	// asset manager is needed to load animation data. we don't load the meshes themselves to avoid
	// depending on OpenGL on the server
	assetManager := assets.NewAssetManager(assetsDirectory, false)
	d.RegisterAssetManager(assetManager)

	playerRegistrationSystem := playerregistration.NewPlayerRegistrationSystem(g, settings.ListenAddress, fmt.Sprintf("%d", settings.Port), settings.ConnectionType)
	networkDispatchSystem := networkdispatch.NewNetworkDispatchSystem(g)
	rpcReceiverSystem := rpcreceiver.NewRPCReceiverSystem(g)
	playerInputSystem := playerinput.NewPlayerInputSystem(g)
	aiSystem := ai.NewAnimationSystem(g)
	preframeSystem := preframe.NewPreFrameSystem(g)

	// systems that can manipulate the transform of an entity
	characterControllerSystem := charactercontroller.NewCharacterControllerSystem(g)
	physicsSystem := physics.NewPhysicsSystem(g)
	collisionSystem := collision.NewCollisionSystem(g)

	abilitySystem := ability.NewAbilitySystem(g)
	combatSystem := combat.NewCombatSystem(g)
	lootSystem := loot.NewLootSystem(g)
	animationSystem := animation.NewAnimationSystem(g)
	networkUpdateSystem := networkupdate.NewNetworkUpdateSystem(g)
	bookKeepingSystem := bookkeeping.NewBookKeepingSystem(g)

	g.systems = append(g.systems, []System{
		playerRegistrationSystem,
		networkDispatchSystem,
		playerInputSystem,
		rpcReceiverSystem,
		aiSystem,
		preframeSystem,
		characterControllerSystem,
		physicsSystem,
		collisionSystem,
		abilitySystem,
		combatSystem,
		lootSystem,
		animationSystem,
		bookKeepingSystem,
		networkUpdateSystem,
	}...)
}
