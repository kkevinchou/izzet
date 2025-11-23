package serialization

import (
	"encoding/json"

	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/geometry"
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

func initDeserializedEntity(e *entity.Entity, assetManager *assets.AssetManager) {
	// set dirty flags
	e.DirtyTransformFlag = true

	// rebuild animation player
	if e.Animation != nil {
		handle := e.Animation.AnimationHandle
		e.Animation = entity.NewAnimationComponent(handle, assetManager)
	}

	if e.MeshComponent != nil && e.Collider != nil {
		// kinda hacky, but right now we only support one collider type per entity.
		// only if all other colliders aren't present do we construct a tri mesh collider (bounding box being the exception)
		if e.Collider.CapsuleCollider == nil {
			// rebuild trimesh collider
			meshHandle := e.MeshComponent.MeshHandle
			primitives := assetManager.GetPrimitives(meshHandle)
			if len(primitives) > 0 {
				primitives := assetManager.GetPrimitives(meshHandle)
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
