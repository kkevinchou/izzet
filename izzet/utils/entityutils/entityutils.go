package entityutils

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/knetwork"
	"github.com/kkevinchou/izzet/izzet/types"
)

func Spawn(entityType types.EntityType, position mgl64.Vec3, orientation mgl64.Quat) *entities.EntityImpl {
	var newEntity *entities.EntityImpl

	if types.EntityType(entityType) == types.EntityTypeBob {
		newEntity = entities.NewBob()
	} else if types.EntityType(entityType) == types.EntityTypeScene {
		newEntity = entities.NewScene()
	} else if types.EntityType(entityType) == types.EntityTypeDynamicRigidBody {
		newEntity = entities.NewDynamicRigidBody()
	} else if types.EntityType(entityType) == types.EntityTypeStaticRigidBody {
		newEntity = entities.NewStaticRigidBody()
	} else if types.EntityType(entityType) == types.EntityTypeProjectile {
		newEntity = entities.NewProjectile(position)
	} else if types.EntityType(entityType) == types.EntityTypeEnemy {
		newEntity = entities.NewEnemy()
	} else if types.EntityType(entityType) == types.EntityTypeLootbox {
		newEntity = entities.NewLootbox()
	} else {
		fmt.Printf("unrecognized entity with type %v to spawn\n", entityType)
		return nil
	}

	cc := newEntity.GetComponentContainer()
	cc.TransformComponent.Position = position
	cc.TransformComponent.Orientation = orientation
	return newEntity
}

func SpawnWithID(entityID int, entityType types.EntityType, position mgl64.Vec3, orientation mgl64.Quat) *entities.EntityImpl {
	newEntity := Spawn(entityType, position, orientation)
	newEntity.ID = entityID
	return newEntity
}

func ConstructEntitySnapshot(entity entities.Entity) knetwork.EntitySnapshot {
	cc := entity.GetComponentContainer()
	transformComponent := cc.TransformComponent
	tpcComponent := cc.ThirdPersonControllerComponent

	snapshot := knetwork.EntitySnapshot{
		ID:          entity.GetID(),
		Type:        int(entity.Type()),
		Position:    transformComponent.Position,
		Orientation: transformComponent.Orientation,
		Components:  cc.Serialize(),
	}

	if tpcComponent != nil {
		snapshot.Velocity = tpcComponent.BaseVelocity
	}

	animationComponent := cc.AnimationComponent
	if animationComponent != nil {
		snapshot.Animation = animationComponent.Player.CurrentAnimation()
	}

	return snapshot
}

// methods
// 		SpawnWithSnapshot(snapshot)
//			should call SpawnX() based on the entity type. any extra params for the entity should come from
//			a bytes blob from the snapshot
//		SpawnX() -- X for each entity type
//		ConstructEntitySnapshot(entity entities.Entity)
//
// goals
//		1. replicate an entity between server to client. spawning doesn't always accomplish this if there's dynamic internal state.
//			should this be related to state synchronization? or entity spawning? perhaps spawning should be one step, and state synch another
//			yes, should be different. we will want to be able to synchronize internal state
//		2. spawn an entity from scratch
