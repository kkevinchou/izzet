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
)

type App interface {
	GetPlayers() map[int]network.Player
	RegisterPlayer(playerID int, connection net.Conn) network.Player
	InputBuffer() *inputbuffer.InputBuffer
	CommandFrame() int
	ModelLibrary() *modellibrary.ModelLibrary
	GetPlayer(playerID int) network.Player
}

type EventsSystem struct {
	app        App
	serializer *serialization.Serializer
}

func NewEventsSystem(app App, serializer *serialization.Serializer) *EventsSystem {
	return &EventsSystem{app: app, serializer: serializer}
}

func (s *EventsSystem) Update(delta time.Duration, world systems.GameWorld) {
	// players := s.app.GetPlayers()
	for _, event := range world.GetEvents() {
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
				CollisionMask: entities.ColliderGroupFlagTerrain,
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

			message, err := createEntityMessage(e.PlayerID, entity)
			if err != nil {
				panic(err)
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				panic(err)
			}
			player.Connection.Write(messageBytes)

			message, err = createEntityMessage(e.PlayerID, camera)
			if err != nil {
				panic(err)
			}

			messageBytes, err = json.Marshal(message)
			if err != nil {
				panic(err)
			}
			player.Connection.Write(messageBytes)

			fmt.Printf("player %d joined, camera %d, entityID %d\n", e.PlayerID, camera.GetID(), entity.GetID())
		}
	}
	world.ClearEventQueue()
}

func createCamera(playerID int, targetEntityID int) *entities.Entity {
	entity := entities.InstantiateEntity("camera")
	entity.CameraComponent = &entities.CameraComponent{TargetPositionOffset: mgl64.Vec3{0, 50, 0}, Target: &targetEntityID}
	entity.ImageInfo = entities.NewImageInfo("camera.png", 15)
	entity.Billboard = true
	entity.PlayerInput = &entities.PlayerInputComponent{PlayerID: playerID}
	return entity
}

func createEntityMessage(playerID int, entity *entities.Entity) (network.Message, error) {
	createEntityMessage := network.CreateEntityMessage{
		OwnerID: playerID,
	}

	cameraBytes, err := json.Marshal(entity)
	if err != nil {
		return network.Message{}, err
	}
	createEntityMessage.EntityBytes = cameraBytes

	bytes, err := json.Marshal(createEntityMessage)
	if err != nil {
		panic(err)
	}

	return network.Message{MessageType: network.MsgTypeCreateEntity, Timestamp: time.Now(), Body: bytes}, nil
}
