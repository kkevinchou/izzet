package serversystem

import (
	"bytes"
	"time"

	"github.com/kkevinchou/izzet/internal/iztlog"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/prefab"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/system"
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

		playerEntity := prefab.CreatePlayer(s.app)
		spawnPoint := world.GetSpawnPoint()
		if spawnPoint != nil {
			entity.SetLocalPosition(playerEntity, spawnPoint.Position())
		}

		camera := prefab.CreateCamera(e.PlayerID)
		playerEntity.CharacterControllerComponent.CameraEntityID = camera.GetID()
		camera.CameraComponent.Target = &playerEntity.ID

		world.AddEntity(playerEntity)
		world.AddEntity(camera)

		message, err := createAckPlayerJoinMessage(e.PlayerID, camera.ID, playerEntity.ID, s.app.World(), s.app.ProjectName())
		if err != nil {
			panic(err)
		}

		err = player.Client.Send(message, s.app.CommandFrame())
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
		world.AddEntity(e.Entity)
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

func createAckPlayerJoinMessage(playerID int, cameraEntityID int, playerEntityID int, world *world.GameWorld, projectName string) (network.AckPlayerJoinMessage, error) {
	ackPlayerJoinMessage := network.AckPlayerJoinMessage{PlayerID: playerID, ProjectName: projectName}

	var worldBytesBuffer bytes.Buffer
	serialization.Write(world, &worldBytesBuffer)

	ackPlayerJoinMessage.PlayerEntityID = playerEntityID
	ackPlayerJoinMessage.CameraEntityID = cameraEntityID
	ackPlayerJoinMessage.SerializedWorld = worldBytesBuffer.Bytes()

	return ackPlayerJoinMessage, nil
}
