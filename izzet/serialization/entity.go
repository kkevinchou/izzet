package serialization

import (
	"encoding/json"

	iztanimation "github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
)

func SerializeEntity(e *entity.Entity) ([]byte, error) {
	return json.Marshal(e)
}

func DeserializeEntity(bytes []byte, assetManager *assets.AssetManager) (*entity.Entity, error) {
	var e entity.Entity
	err := json.Unmarshal(bytes, &e)
	if err != nil {
		return nil, err
	}

	initDeserializedEntity(&e, assetManager)
	return &e, err
}

func initDeserializedEntity(e *entity.Entity, am *assets.AssetManager) {
	e.DirtyTransformFlag = true

	// reinitialize the animation player and state machine
	if e.Animation != nil {
		animations, joints, rootJointID := am.GetAnimations(e.Animation.AnimationHandle)
		e.Animation.AnimationPlayer = iztanimation.NewAnimationPlayer()
		e.Animation.AnimationPlayer.Initialize(animations, joints[rootJointID])

		if e.Animation.AnimationStateMachine != nil {
			currentState := e.Animation.AnimationStateMachine.CurrentState.Name
			e.Animation.AnimationStateMachine = animation.NewStateMachine(e.Animation.AnimationStateMachineID)
			e.Animation.AnimationStateMachine.SetCurrentState(currentState)
			e.Animation.AnimationPlayer.PlayClip(e.Animation.AnimationStateMachine.CurrentState.ClipName)
		}
	}

	if e.MeshComponent != nil && e.Collider != nil {
		// kinda hacky, but right now we only support one collider type per entity.
		// only if all other colliders aren't present do we construct a tri mesh collider (bounding box being the exception)
		if e.Collider.CapsuleCollider == nil {
			// rebuild trimesh collider
			meshHandle := e.MeshComponent.MeshHandle
			primitives := am.GetPrimitives(meshHandle)
			if len(primitives) > 0 {
				primitives := am.GetPrimitives(meshHandle)
				t := collider.CreateTriMeshFromPrimitives(entity.AssetPrimitiveToSpecPrimitive(primitives))
				bb := collider.BoundingBoxFromVertices(assets.UniqueVerticesFromPrimitives(primitives))
				var simplifiedTriMesh *collider.TriMesh
				if e.SimplifiedTriMeshIterations > 0 {
					simplifiedTriMesh = geometry.SimplifyMesh(entity.AssetPrimitiveToSpecPrimitive(primitives)[0], e.SimplifiedTriMeshIterations)
				}
				e.Collider = entity.CreateTriMeshColliderComponent(e.Collider.ColliderGroup, 0, *t, simplifiedTriMesh, bb)
			}
		} else {
			e.Collider = entity.CreateCapsuleColliderComponent(e.Collider.ColliderGroup, e.Collider.CollisionMask, *e.Collider.CapsuleCollider)
		}
	}
}
