package serversystems

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/izzet/izzet/systems"
	"github.com/kkevinchou/izzet/izzet/types"
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

		if e.AIComponent.PathfindConfig == nil {
			continue
		}
		e.AIComponent.PathfindConfig.Goal = rpc.Pathfind.Goal
		e.AIComponent.PathfindConfig.State = entity.PathfindingStateGoalSet
	}
}

func (s *ReceiverSystem) handleCreateEntityRPC(rpc network.RPCMessage) {
	var modelName string

	// TODO: this should probably imported as a data file rather than hard coded
	var idleAnimation string
	var attackAnimation string
	var runAnimation string
	if rpc.CreateEntity.EntityType == string(entity.EntityTypeVelociraptor) {
		modelName = "velociraptor"
		idleAnimation = "Velociraptor_Idle"
		attackAnimation = "Velociraptor_Attack"
		runAnimation = "Velociraptor_Run"
	} else if rpc.CreateEntity.EntityType == string(entity.EntityTypeParasaurolophus) {
		modelName = "parasaurolophus"
		idleAnimation = "Parasaurolophus_Idle"
		attackAnimation = "Parasaurolophus_Attack"
		runAnimation = "Parasaurolophus_Run"
	}

	handle := assets.NewSingleEntityMeshHandle(modelName)
	e := entity.CreateEmptyEntity(modelName)
	e.Kinematic = &entity.KinematicComponent{GravityEnabled: true}

	capsule := collider.NewCapsule(mgl64.Vec3{0, 3, 0}, mgl64.Vec3{0, 1, 0}, 1)
	e.Collider = entity.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
	e.Collider.CapsuleCollider = &capsule

	e.MeshComponent = &entity.MeshComponent{MeshHandle: handle, Transform: mgl64.Rotate3DY(180 * math.Pi / 180).Mat4(), Visible: true, ShadowCasting: true}
	e.Animation = entity.NewAnimationComponent(modelName, s.app.AssetManager())
	e.Animation.AnimationNames[entity.AnimationKeyIdle] = idleAnimation
	e.Animation.AnimationNames[entity.AnimationKeyAttack] = attackAnimation
	e.Animation.AnimationNames[entity.AnimationKeyRun] = runAnimation

	jitterX := rand.Intn(10)
	jitterZ := rand.Intn(10)
	entity.SetLocalPosition(e, mgl64.Vec3{float64(jitterX), 20, float64(jitterZ)})
	entity.SetScale(e, mgl64.Vec3{0.5, 0.5, 0.5})

	e.AIComponent = &entity.AIComponent{
		Speed: 7,
		// AttackConfig:   &entity.AttackConfig{},
	}

	if rpc.CreateEntity.Patrol {
		targetDist := 20
		jitterTargetX := rand.Intn(targetDist) - 10
		jitterTargetZ := rand.Intn(targetDist) - 10
		target := mgl64.Vec3{float64(jitterTargetX), 0, float64(jitterTargetZ)}.Normalize().Mul(float64(targetDist))
		e.AIComponent.PatrolConfig = &entity.PatrolConfig{Points: []mgl64.Vec3{{float64(jitterX), 0, float64(jitterZ)}, target}}
	} else {
		e.AIComponent.PathfindConfig = &entity.PathfindConfig{}

	}

	world := s.app.World()
	for _, spawnPoint := range world.Entities() {
		if spawnPoint.SpawnPointComponent != nil {
			entity.SetLocalPosition(e, spawnPoint.Position())
			// entity.SetLocalPosition(e, mgl64.Vec3{spawnPoint.Position().X() + float64(jitterX), spawnPoint.Position().Y(), spawnPoint.Position().Z() + float64(jitterZ)})
			break
		}
	}

	s.app.EventsManager().EntitySpawnTopic.Write(events.EntitySpawnEvent{Entity: e})
}
