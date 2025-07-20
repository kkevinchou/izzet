package serversystems

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/collision/collider"
)

type EventsSystem struct {
	app                      App
	playerJoinConsumer       *events.Consumer[events.PlayerJoinEvent]
	playerDisconnectConsumer *events.Consumer[events.PlayerDisconnectEvent]
	entitySpawnConsumer      *events.Consumer[events.EntitySpawnEvent]
}

func NewEventsSystem(app App) *EventsSystem {
	eventsManager := app.EventsManager()
	return &EventsSystem{
		app:                      app,
		playerJoinConsumer:       events.NewConsumer(eventsManager.PlayerJoinTopic),
		playerDisconnectConsumer: events.NewConsumer(eventsManager.PlayerDisconnectTopic),
		entitySpawnConsumer:      events.NewConsumer(eventsManager.EntitySpawnTopic),
	}
}

func (s *EventsSystem) Name() string {
	return "EventsSystem"
}

func (s *EventsSystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, e := range s.playerJoinConsumer.ReadNewEvents() {
		player := s.app.RegisterPlayer(e.PlayerID, e.Connection)

		var radius float64 = 40
		var length float64 = 80
		entity := entities.CreateEmptyEntity("player")
		entity.PositionSync = &entities.PositionSync{}
		// entity.Physics = &entities.PhysicsComponent{GravityEnabled: true}
		entity.Kinematic = &entities.KinematicComponent{GravityEnabled: true}
		capsule := collider.Capsule{
			Radius: radius,
			Top:    mgl64.Vec3{0, radius + length, 0},
			Bottom: mgl64.Vec3{0, radius, 0},
		}
		entity.Collider = entities.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
		entity.CharacterControllerComponent = &entities.CharacterControllerComponent{Speed: settings.CharacterSpeed, FlySpeed: settings.CharacterFlySpeed}
		handle := assets.NewSingleEntityMeshHandle("alpha3")

		entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true, InvisibleToPlayerOwner: settings.FirstPersonCamera}
		entity.Animation = entities.NewAnimationComponent("alpha3", s.app.AssetManager())
		entity.RenderBlend = &entities.RenderBlend{}
		entities.SetScale(entity, mgl64.Vec3{0.01, 0.01, 0.01})

		world := s.app.World()
		for _, e := range world.Entities() {
			if e.SpawnPointComponent != nil {
				entities.SetLocalPosition(entity, e.Position())
				break
			}
		}

		camera := createCamera(e.PlayerID, entity.GetID())
		world.AddEntity(camera)
		world.AddEntity(entity)

		worldBytes := s.app.SerializeWorld()
		message, err := createAckPlayerJoinMessage(e.PlayerID, camera, entity, worldBytes, s.app.ProjectName())
		if err != nil {
			panic(err)
		}
		player.Client.Send(message, s.app.CommandFrame())
		if err != nil {
			panic(err)
		}

		s.spawnEntity(world, camera)
		s.spawnEntity(world, entity)
		fmt.Printf("player %d joined, camera %d, entityID %d\n", e.PlayerID, camera.GetID(), entity.GetID())
	}

	for _, e := range s.playerDisconnectConsumer.ReadNewEvents() {
		fmt.Printf("player %d disconnected\n", e.PlayerID)
		s.app.DeregisterPlayer(e.PlayerID)
	}

	for _, e := range s.entitySpawnConsumer.ReadNewEvents() {
		s.spawnEntity(world, e.Entity)
	}
}

func (s *EventsSystem) spawnEntity(world systems.GameWorld, entity *entities.Entity) {
	world.AddEntity(entity)
	entityMessage, err := createEntityMessage(0, entity)
	if err != nil {
		panic(err)
	}
	for _, player := range s.app.GetPlayers() {
		player.Client.Send(entityMessage, s.app.CommandFrame())
	}
	fmt.Printf("spawned entity with ID %d\n", entity.GetID())
}

func createEntityMessage(playerID int, entity *entities.Entity) (network.CreateEntityMessage, error) {
	createEntityMessage := network.CreateEntityMessage{
		OwnerID: playerID,
	}

	entityBytes, err := json.Marshal(entity)
	if err != nil {
		return network.CreateEntityMessage{}, err
	}
	createEntityMessage.EntityBytes = entityBytes

	return createEntityMessage, nil
}

func createCamera(playerID int, targetEntityID int) *entities.Entity {
	entity := entities.CreateEmptyEntity("camera")
	entity.CameraComponent = &entities.CameraComponent{TargetPositionOffset: mgl64.Vec3{0, settings.CameraEntityFollowVerticalOffset, 0}, Target: &targetEntityID}
	entity.ImageInfo = entities.NewImageInfo("camera.png", 1)
	entity.Billboard = true
	entity.PlayerInput = &entities.PlayerInputComponent{PlayerID: playerID}
	return entity
}

func createAckPlayerJoinMessage(playerID int, camera *entities.Entity, entity *entities.Entity, worldBytes []byte, projectName string) (network.AckPlayerJoinMessage, error) {
	ackPlayerJoinMessage := network.AckPlayerJoinMessage{PlayerID: playerID, ProjectName: projectName}

	entityBytes, err := json.Marshal(entity)
	if err != nil {
		return network.AckPlayerJoinMessage{}, err
	}
	cameraBytes, err := json.Marshal(camera)
	if err != nil {
		return network.AckPlayerJoinMessage{}, err
	}

	ackPlayerJoinMessage.EntityBytes = entityBytes
	ackPlayerJoinMessage.CameraBytes = cameraBytes
	ackPlayerJoinMessage.Snapshot = worldBytes

	return ackPlayerJoinMessage, nil
}
