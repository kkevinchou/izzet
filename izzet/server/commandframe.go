package server

import (
	"encoding/json"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/events"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/network"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

func (g *Server) runCommandFrame(delta time.Duration) {
	g.commandFrame += 1
	// THIS NEEDS TO BE THE FIRST THING THAT RUNS TO MAKE SURE THE SPATIAL PARTITION
	// HAS A CHANCE TO SEE THE ENTITY AND INDEX IT
	g.handleSpatialPartition()

	g.handlePlayerConnections()
	for _, s := range g.systems {
		s.Update(delta, g.world)
	}
	g.replicator.Update(delta, g.world)
}

func (g *Server) handlePlayerConnections() {
	select {
	case connection := <-g.newConnections:
		g.world.QueueEvent(events.PlayerJoinEvent{PlayerID: connection.PlayerID, Connection: connection.Connection})
		player := g.RegisterPlayer(connection.PlayerID, connection.Connection)

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
		entity.Animation = entities.NewAnimationComponent("alpha", g.ModelLibrary())
		entities.SetScale(entity, mgl64.Vec3{0.25, 0.25, 0.25})

		camera := createCamera(connection.PlayerID, entity.GetID())
		g.world.AddEntity(camera)
		g.world.AddEntity(entity)

		message, err := createAckPlayerJoinMessage(camera, entity)
		if err != nil {
			panic(err)
		}
		messageBytes, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}
		player.Connection.Write(messageBytes)

		g.world.QueueEvent(events.PlayerJoinEvent{
			PlayerID:       connection.PlayerID,
			Connection:     connection.Connection,
			PlayerEntityID: entity.GetID(),
			PlayerCameraID: camera.GetID(),
		})
	default:
		return
	}
}

func (g *Server) handleSpatialPartition() {
	var spatialEntities []spatialpartition.Entity
	for _, entity := range g.world.Entities() {
		if !entity.HasBoundingBox() {
			continue
		}
		spatialEntities = append(spatialEntities, entity)
	}
	g.world.SpatialPartition().IndexEntities(spatialEntities)
}

func createCamera(playerID int, targetEntityID int) *entities.Entity {
	entity := entities.InstantiateEntity("camera")
	entity.CameraComponent = &entities.CameraComponent{TargetPositionOffset: mgl64.Vec3{0, 50, 0}, Target: &targetEntityID}
	entity.ImageInfo = entities.NewImageInfo("camera.png", 15)
	entity.Billboard = true
	entity.PlayerInput = &entities.PlayerInputComponent{PlayerID: playerID}
	return entity
}

func createAckPlayerJoinMessage(camera *entities.Entity, entity *entities.Entity) (network.Message, error) {
	ackPlayerJoinMessage := network.AckPlayerJoinMessage{}

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

	bytes, err := json.Marshal(ackPlayerJoinMessage)
	if err != nil {
		panic(err)
	}

	return network.Message{MessageType: network.MsgTypeAckPlayerJoin, Timestamp: time.Now(), Body: bytes}, nil
}
