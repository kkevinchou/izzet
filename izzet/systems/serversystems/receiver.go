package serversystems

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/kitolib/collision/collider"
)

type ReceiverSystem struct {
	app App
}

func NewReceiverSystem(app App) *ReceiverSystem {
	return &ReceiverSystem{app: app}
}

func (s *ReceiverSystem) Name() string {
	return "ReceiverSystem"
}

func (s *ReceiverSystem) Update(delta time.Duration, world systems.GameWorld) {
	for _, player := range s.app.GetPlayers() {
		noMessage := false
		for !noMessage {
			select {
			case message := <-player.InMessageChannel:
				if message.MessageType == network.MsgTypePlayerInput {
					inputMessage, err := network.ExtractMessage[network.InputMessage](message)
					if err != nil {
						fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
						continue
					}
					s.app.InputBuffer().PushInput(message.CommandFrame, player.ID, inputMessage.Input)
				} else if message.MessageType == network.MsgTypePing {
					pingMessage, err := network.ExtractMessage[network.PingMessage](message)
					if err != nil {
						fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
						continue
					}
					player.Client.Send(pingMessage, s.app.CommandFrame())
				} else if message.MessageType == network.MsgTypeRPC {
					rpc, err := network.ExtractMessage[network.RPCMessage](message)
					if err != nil {
						fmt.Println(fmt.Errorf("failed to deserialize message %w", err))
						continue
					}

					if rpc.Pathfind != nil {
						s.handlePathfindRPC(rpc)
					}

					if rpc.CreateEntity != nil {
						s.handleCreateEntityRPC(rpc)
					}
				}
			case <-player.DisconnectChannel:
				s.app.EventsManager().PlayerDisconnectTopic.Write(events.PlayerDisconnectEvent{PlayerID: player.ID})
			default:
				noMessage = true
			}
		}
	}
}

func (s *ReceiverSystem) handlePathfindRPC(rpc network.RPCMessage) {
	for _, e := range s.app.World().Entities() {
		if e.AIComponent == nil {
			continue
		}
		e.AIComponent.PathfindConfig.Goal = rpc.Pathfind.Goal
		e.AIComponent.PathfindConfig.State = entities.PathfindingStateGoalSet
	}
}

func (s *ReceiverSystem) handleCreateEntityRPC(rpc network.RPCMessage) {
	var modelName string

	// TODO: this should probably imported as a data file rather than hard coded
	var idleAnimation string
	var attackAnimation string
	var runAnimation string
	if rpc.CreateEntity.EntityType == string(entities.EntityTypeVelociraptor) {
		modelName = "velociraptor"
		idleAnimation = "Velociraptor_Idle"
		attackAnimation = "Velociraptor_Attack"
		runAnimation = "Velociraptor_Run"
	} else if rpc.CreateEntity.EntityType == string(entities.EntityTypeParasaurolophus) {
		modelName = "parasaurolophus"
		idleAnimation = "Parasaurolophus_Idle"
		attackAnimation = "Parasaurolophus_Attack"
		runAnimation = "Parasaurolophus_Run"
	}

	handle := assets.NewGlobalHandle(modelName)
	entity := entities.InstantiateEntity(modelName)
	entity.Physics = &entities.PhysicsComponent{GravityEnabled: true, RotateOnVelocity: false}
	entity.Collider = &entities.ColliderComponent{
		ColliderGroup: entities.ColliderGroupFlagPlayer,
		CollisionMask: entities.ColliderGroupFlagTerrain | entities.ColliderGroupFlagPlayer,
	}

	// primitives := s.app.AssetManager().GetPrimitives(handle)
	// verts := assets.UniqueVerticesFromPrimitives(primitives)
	// c := collider.NewCapsuleFromVertices(verts)
	c := collider.NewCapsule(mgl64.Vec3{0, 4, 0}, mgl64.Vec3{0, 2, 0}, 2)
	entity.Collider.CapsuleCollider = &c

	capsule := entity.Collider.CapsuleCollider
	entity.InternalBoundingBox = collider.BoundingBox{MinVertex: capsule.Bottom.Sub(mgl64.Vec3{c.Radius, c.Radius, c.Radius}), MaxVertex: capsule.Top.Add(mgl64.Vec3{c.Radius, c.Radius, c.Radius})}

	entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true}
	entity.Animation = entities.NewAnimationComponent(modelName, s.app.AssetManager())
	entity.Animation.AnimationNames[entities.AnimationKeyIdle] = idleAnimation
	entity.Animation.AnimationNames[entities.AnimationKeyAttack] = attackAnimation
	entity.Animation.AnimationNames[entities.AnimationKeyRun] = runAnimation

	jitterX := rand.Intn(10)
	jitterZ := rand.Intn(10)
	entities.SetLocalPosition(entity, mgl64.Vec3{float64(jitterX), 60, float64(jitterZ)})

	entity.AIComponent = &entities.AIComponent{
		Speed:          25,
		PathfindConfig: &entities.PathfindConfig{},
		// AttackConfig:   &entities.AttackConfig{},
	}

	world := s.app.World()
	for _, e := range world.Entities() {
		if e.SpawnPointComponent != nil {
			entities.SetLocalPosition(entity, e.Position())
			entities.SetLocalPosition(entity, mgl64.Vec3{e.Position().X() + float64(jitterX), 60, e.Position().Z() + float64(jitterZ)})
			break
		}
	}

	s.app.EventsManager().EntitySpawnTopic.Write(events.EntitySpawnEvent{Entity: entity})
}
