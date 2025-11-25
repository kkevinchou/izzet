package serversystems

import (
	"bytes"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/iztlog"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/system"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/izzet/izzet/world"
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

func (s *EventsSystem) Update(delta time.Duration, world system.GameWorld) {
	for _, e := range s.playerJoinConsumer.ReadNewEvents() {
		player := s.app.RegisterPlayer(e.PlayerID, e.Connection)

		var radius float64 = 40
		var length float64 = 80
		playerEntity := entity.CreateEmptyEntity("player")
		playerEntity.Kinematic = &entity.KinematicComponent{GravityEnabled: true}
		capsule := collider.Capsule{
			Radius: radius,
			Top:    mgl64.Vec3{0, radius + length, 0},
			Bottom: mgl64.Vec3{0, radius, 0},
		}
		playerEntity.Collider = entity.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
		playerEntity.CharacterControllerComponent = &entity.CharacterControllerComponent{Speed: settings.CharacterSpeed, FlySpeed: settings.CharacterFlySpeed}
		handle := assets.NewSingleEntityMeshHandle("alpha3")

		playerEntity.MeshComponent = &entity.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true, InvisibleToPlayerOwner: settings.FirstPersonCamera}
		playerEntity.Animation = entity.NewAnimationComponent("alpha3", s.app.AssetManager())
		playerEntity.RenderBlend = &entity.RenderBlend{}
		entity.SetScale(playerEntity, mgl64.Vec3{0.01, 0.01, 0.01})

		world := s.app.World()
		for _, spawnPoint := range world.Entities() {
			if spawnPoint.SpawnPointComponent != nil {
				entity.SetLocalPosition(playerEntity, spawnPoint.Position())
				break
			}
		}

		camera := createCamera(e.PlayerID, playerEntity.GetID())

		// technically spawnEntity will also attempt to add the entity to the game world
		// but we want to pre-emptively add the entities before serializing the player ack
		// message
		world.AddEntity(playerEntity)
		world.AddEntity(camera)

		message, err := createAckPlayerJoinMessage(e.PlayerID, camera.ID, playerEntity.ID, s.app.World(), s.app.ProjectName())
		if err != nil {
			panic(err)
		}
		player.Client.Send(message, s.app.CommandFrame())
		if err != nil {
			panic(err)
		}

		s.notifyEntityCreation(camera)
		s.notifyEntityCreation(playerEntity)
		iztlog.Logger.Info("player joined", "player id", e.PlayerID, "camera id", camera.GetID(), "player entity id", playerEntity.GetID())
	}

	for _, e := range s.playerDisconnectConsumer.ReadNewEvents() {
		iztlog.Logger.Info("player disconnected", "player id", e.PlayerID)
		s.app.DeregisterPlayer(e.PlayerID)
	}

	for _, e := range s.entitySpawnConsumer.ReadNewEvents() {
		s.notifyEntityCreation(e.Entity)
	}
}

func (s *EventsSystem) notifyEntityCreation(entity *entity.Entity) {
	entityMessage, err := createEntityMessage(0, entity)
	if err != nil {
		panic(err)
	}
	for _, player := range s.app.GetPlayers() {
		player.Client.Send(entityMessage, s.app.CommandFrame())
	}
	iztlog.Logger.Info("spawned entity", "entity id", entity.GetID())
}

func createEntityMessage(playerID int, entity *entity.Entity) (network.CreateEntityMessage, error) {
	createEntityMessage := network.CreateEntityMessage{
		OwnerID: playerID,
	}

	entityBytes, err := serialization.SerializeEntity(entity)
	if err != nil {
		return network.CreateEntityMessage{}, err
	}
	createEntityMessage.EntityBytes = entityBytes

	return createEntityMessage, nil
}

func createCamera(playerID int, targetEntityID int) *entity.Entity {
	e := entity.CreateEmptyEntity("camera")
	e.CameraComponent = &entity.CameraComponent{TargetPositionOffset: mgl64.Vec3{0, settings.CameraEntityFollowVerticalOffset, 0}, Target: &targetEntityID}
	e.ImageInfo = entity.NewImageInfo("camera.png", 1)
	e.Billboard = true
	e.PlayerInput = &entity.PlayerInputComponent{PlayerID: playerID}
	return e
}

func createAckPlayerJoinMessage(playerID int, cameraEntityID int, playerEntityID int, world *world.GameWorld, projectName string) (network.AckPlayerJoinMessage, error) {
	ackPlayerJoinMessage := network.AckPlayerJoinMessage{PlayerID: playerID, ProjectName: projectName}

	var worldBytesBuffer bytes.Buffer
	serialization.Write(world, &worldBytesBuffer)

	ackPlayerJoinMessage.PlayerEntityID = playerEntityID
	ackPlayerJoinMessage.CameraEntityID = cameraEntityID
	ackPlayerJoinMessage.SerializedWorld = worldBytesBuffer.Bytes()

	return ackPlayerJoinMessage, nil
}
