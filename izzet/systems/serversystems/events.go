package serversystems

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/server/inputbuffer"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/input"
)

type App interface {
	GetPlayers() map[int]*network.Player
	RegisterPlayer(playerID int, connection net.Conn) *network.Player
	InputBuffer() *inputbuffer.InputBuffer
	CommandFrame() int
	ModelLibrary() *modellibrary.ModelLibrary
	GetPlayer(playerID int) *network.Player
	GetPlayerInput(playerID int) input.Input
	SetPlayerInput(playerID int, input input.Input)
	DeregisterPlayer(playerID int)
	SerializeWorld() []byte
}

type EventsSystem struct {
	app        App
	serializer *serialization.Serializer
}

func NewEventsSystem(app App, serializer *serialization.Serializer) *EventsSystem {
	return &EventsSystem{app: app, serializer: serializer}
}

func (s *EventsSystem) Update(delta time.Duration, world systems.GameWorld) {
	worldEvents := world.GetEvents()
	var nextEventIndex = 0

	for nextEventIndex < len(worldEvents) {
		event := worldEvents[nextEventIndex]
		switch e := event.(type) {
		case events.PlayerJoinEvent:
			player := s.app.RegisterPlayer(e.PlayerID, e.Connection)

			var radius float64 = 40
			var length float64 = 80
			entity := entities.InstantiateEntity("player")
			entity.Physics = &entities.PhysicsComponent{GravityEnabled: true}
			entity.Collider = &entities.ColliderComponent{
				CapsuleCollider: &collider.Capsule{
					Radius: radius,
					Top:    mgl64.Vec3{0, radius + length, 0},
					Bottom: mgl64.Vec3{0, radius, 0},
				},
				ColliderGroup: entities.ColliderGroupFlagPlayer,
				CollisionMask: entities.ColliderGroupFlagTerrain | entities.ColliderGroupFlagPlayer,
			}
			entity.CharacterControllerComponent = &entities.CharacterControllerComponent{Speed: 100}

			capsule := entity.Collider.CapsuleCollider
			entity.InternalBoundingBox = collider.BoundingBox{MinVertex: capsule.Bottom.Sub(mgl64.Vec3{radius, radius, radius}), MaxVertex: capsule.Top.Add(mgl64.Vec3{radius, radius, radius})}

			handle := modellibrary.NewGlobalHandle("alpha")
			entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4()}
			entity.Animation = entities.NewAnimationComponent("alpha", s.app.ModelLibrary())
			entities.SetScale(entity, mgl64.Vec3{0.25, 0.25, 0.25})

			camera := createCamera(e.PlayerID, entity.GetID())
			world.AddEntity(camera)
			world.AddEntity(entity)

			worldBytes := s.app.SerializeWorld()
			message, err := createAckPlayerJoinMessage(e.PlayerID, camera, entity, worldBytes)
			if err != nil {
				panic(err)
			}
			messageBytes, err := json.Marshal(message)
			if err != nil {
				panic(err)
			}
			player.Connection.Write(messageBytes)

			world.QueueEvent(events.EntitySpawnEvent{Entity: camera})
			world.QueueEvent(events.EntitySpawnEvent{Entity: entity})
			fmt.Printf("player %d joined, camera %d, entityID %d\n", e.PlayerID, camera.GetID(), entity.GetID())
		case events.PlayerDisconnectEvent:
			fmt.Printf("player %d disconnected\n", e.PlayerID)
			s.app.DeregisterPlayer(e.PlayerID)
		case events.EntitySpawnEvent:
			world.AddEntity(e.Entity)
			entityMessage, err := createEntityMessage(0, e.Entity)
			if err != nil {
				panic(err)
			}
			for _, player := range s.app.GetPlayers() {
				player.Client.Send(entityMessage, s.app.CommandFrame())
			}
			fmt.Printf("spawned entity with ID %d\n", e.Entity.GetID())
		default:
		}
		nextEventIndex += 1
		worldEvents = world.GetEvents()
	}

	world.ClearEventQueue()
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
	entity := entities.InstantiateEntity("camera")
	entity.CameraComponent = &entities.CameraComponent{TargetPositionOffset: mgl64.Vec3{0, 50, 0}, Target: &targetEntityID}
	entity.ImageInfo = entities.NewImageInfo("camera.png", 15)
	entity.Billboard = true
	entity.PlayerInput = &entities.PlayerInputComponent{PlayerID: playerID}
	return entity
}

func createAckPlayerJoinMessage(playerID int, camera *entities.Entity, entity *entities.Entity, worldBytes []byte) (network.MessageTransport, error) {
	ackPlayerJoinMessage := network.AckPlayerJoinMessage{PlayerID: playerID}

	entityBytes, err := json.Marshal(entity)
	if err != nil {
		panic(err)
	}
	ackPlayerJoinMessage.EntityBytes = entityBytes

	cameraBytes, err := json.Marshal(camera)
	if err != nil {
		panic(err)
	}
	ackPlayerJoinMessage.CameraBytes = cameraBytes

	ackPlayerJoinMessage.Snapshot = worldBytes

	bytes, err := json.Marshal(ackPlayerJoinMessage)
	if err != nil {
		panic(err)
	}

	return network.MessageTransport{MessageType: network.MsgTypeAckPlayerJoin, Timestamp: time.Now(), Body: bytes}, nil
}
