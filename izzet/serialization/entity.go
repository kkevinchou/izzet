package serialization

import (
	"encoding/json"

	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
)

func SerializeEntity(entity *entities.Entity) ([]byte, error) {
	return json.Marshal(entity)
}

func DeserializeEntity(bytes []byte, assetManager *assets.AssetManager) (*entities.Entity, error) {
	var e entities.Entity
	err := json.Unmarshal(bytes, &e)
	if err != nil {
		return nil, err
	}

	initDeserializedEntity(&e, assetManager)
	return &e, err
}

func initDeserializedEntity(entity *entities.Entity, assetManager *assets.AssetManager) {
	// set dirty flags
	entity.DirtyTransformFlag = true

	// rebuild animation player
	if entity.Animation != nil {
		handle := entity.Animation.AnimationHandle
		entity.Animation = entities.NewAnimationComponent(handle, assetManager)
	}

	if entity.MeshComponent != nil && entity.Collider != nil {
		// kinda hacky, but right now we only support one collider type per entity.
		// only if all other colliders aren't present do we construct a tri mesh collider (bounding box being the exception)
		if entity.Collider.CapsuleCollider == nil {
			// rebuild trimesh collider
			meshHandle := entity.MeshComponent.MeshHandle
			primitives := assetManager.GetPrimitives(meshHandle)
			if len(primitives) > 0 {
				primitives := assetManager.GetPrimitives(meshHandle)
				t := collider.CreateTriMeshFromPrimitives(entities.AssetPrimitiveToSpecPrimitive(primitives))
				bb := collider.BoundingBoxFromVertices(assets.UniqueVerticesFromPrimitives(primitives))
				var simplifiedTriMesh *collider.TriMesh
				if entity.SimplifiedTriMeshIterations > 0 {
					simplifiedTriMesh = geometry.SimplifyMesh(entities.AssetPrimitiveToSpecPrimitive(primitives)[0], entity.SimplifiedTriMeshIterations)
				}
				entity.Collider = entities.CreateTriMeshColliderComponent(entity.Collider.ColliderGroup, 0, *t, simplifiedTriMesh, bb)
			}
		} else {
			entity.Collider = entities.CreateCapsuleColliderComponent(entity.Collider.ColliderGroup, entity.Collider.CollisionMask, *entity.Collider.CapsuleCollider)
		}
	}
}
