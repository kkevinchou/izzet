package serialization

import (
	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision/collider"
)

func InitDeserializedEntity(entity *entities.Entity, ml *assets.AssetManager) {
	// set dirty flags
	entity.DirtyTransformFlag = true

	// rebuild animation player
	if entity.Animation != nil {
		handle := entity.Animation.AnimationHandle
		entity.Animation = entities.NewAnimationComponent(handle, ml)
	}

	if entity.MeshComponent != nil && entity.Collider != nil {
		// kinda hacky, but right now we only support one collider type per entity.
		// only if all other colliders aren't present do we construct a tri mesh collider
		if entity.Collider.CapsuleCollider == nil {
			// rebuild trimesh collider
			meshHandle := entity.MeshComponent.MeshHandle
			primitives := ml.GetPrimitives(meshHandle)
			if len(primitives) > 0 {
				entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(entities.AssetPrimitiveToSpecPrimitive(primitives))
				if entity.SimplifiedTriMeshIterations > 0 {
					entity.Collider.SimplifiedTriMeshCollider = geometry.SimplifyMesh(entities.AssetPrimitiveToSpecPrimitive(primitives)[0], entity.SimplifiedTriMeshIterations)
				}
			}
		}
	}
}

func InitDeserializedEntities(entitySlice []*entities.Entity, ml *assets.AssetManager) {
	for _, entity := range entitySlice {
		InitDeserializedEntity(entity, ml)
	}
}
