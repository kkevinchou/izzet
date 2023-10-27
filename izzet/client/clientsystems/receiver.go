package clientsystems

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/serialization"
	"github.com/kkevinchou/izzet/izzet/systems"
)

type ReceiverSystem struct {
	app App
}

func NewReceiverSystem(app App) *ReceiverSystem {
	return &ReceiverSystem{app: app}
}

func (s *ReceiverSystem) Update(delta time.Duration, world systems.GameWorld) {
	for {
		select {
		case message := <-s.app.NetworkMessagesChannel():
			if message.MessageType == network.MsgTypeGameStateUpdate {
				var gameStateUpdateMessage network.GameStateUpdateMessage
				err := json.Unmarshal(message.Body, &gameStateUpdateMessage)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
					continue
				}

				for _, transform := range gameStateUpdateMessage.Transforms {
					entity := world.GetEntityByID(transform.EntityID)
					if entity == nil {
						continue
					}

					if entity.CameraComponent != nil {
						if entity.PlayerInput.PlayerID == s.app.GetPlayerID() {
							// don't synchronize local camera position
							continue
						}
					}

					entities.SetLocalPosition(entity, transform.Position)
				}
			} else if message.MessageType == network.MsgTypeCreateEntity {
				var createEntityMessage network.CreateEntityMessage
				err := json.Unmarshal(message.Body, &createEntityMessage)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
					continue
				}
				fmt.Println(len(createEntityMessage.EntityBytes))

				var entity entities.Entity
				err = json.Unmarshal(createEntityMessage.EntityBytes, &entity)
				if err != nil {
					fmt.Println(fmt.Errorf("failed to deserialize entity %w", err))
					continue
				}

				fmt.Println("RECEIVED CREATE", entity.ID)

				serialization.InitDeserializedEntity(&entity, s.app.ModelLibrary(), false)
				world.AddEntity(&entity)

				// var radius float64 = 40
				// var length float64 = 80
				// entity := entities.InstantiateEntity("player")
				// entity.Physics = &entities.PhysicsComponent{GravityEnabled: true}
				// entity.Collider = &entities.ColliderComponent{
				// 	CapsuleCollider: &collider.Capsule{
				// 		Radius: radius,
				// 		Top:    mgl64.Vec3{0, radius + length, 0},
				// 		Bottom: mgl64.Vec3{0, radius, 0},
				// 	},
				// 	ColliderGroup: entities.ColliderGroupFlagPlayer,
				// 	CollisionMask: entities.ColliderGroupFlagTerrain,
				// }
				// entity.CharacterControllerComponent = &entities.CharacterControllerComponent{Speed: 100}

				// capsule := entity.Collider.CapsuleCollider
				// entity.InternalBoundingBox = collider.BoundingBox{MinVertex: capsule.Bottom.Sub(mgl64.Vec3{radius, radius, radius}), MaxVertex: capsule.Top.Add(mgl64.Vec3{radius, radius, radius})}

				// handle := modellibrary.NewGlobalHandle("alpha")
				// entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4()}
				// entity.Animation = entities.NewAnimationComponent("alpha", s.app.ModelLibrary())
				// entities.SetScale(entity, mgl64.Vec3{0.25, 0.25, 0.25})

				// camera := createCamera(createEntityMessage.OwnerID, entity.GetID())
				// world.AddEntity(camera)
				// world.AddEntity(entity)
			}
		default:
			return
		}
	}
}

func createCamera(playerID int, targetEntityID int) *entities.Entity {
	entity := entities.InstantiateEntity("camera")
	entity.CameraComponent = &entities.CameraComponent{TargetPositionOffset: mgl64.Vec3{0, 50, 0}, Target: &targetEntityID}
	entity.ImageInfo = entities.NewImageInfo("camera.png", 15)
	entity.Billboard = true
	entity.PlayerInput = &entities.PlayerInputComponent{PlayerID: playerID}
	return entity
}
