package serialization

import (
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/kitolib/collision/collider"
)

func InitDeserializedEntity(entity *entities.Entity, ml *modellibrary.ModelLibrary) {
	// set dirty flags
	entity.DirtyTransformFlag = true

	// rebuild animation player
	if entity.Animation != nil {
		handle := entity.Animation.AnimationHandle
		entity.Animation = entities.NewAnimationComponent(handle, ml)
	}

	// rebuild trimesh collider
	if entity.MeshComponent != nil && entity.Collider != nil {
		meshHandle := entity.MeshComponent.MeshHandle
		primitives := ml.GetPrimitives(meshHandle)
		if len(primitives) > 0 {
			entity.Collider.TriMeshCollider = collider.CreateTriMeshFromPrimitives(entities.MLPrimitivesTospecPrimitive(primitives))
		}
	}
}

func InitDeserializedEntities(entitySlice []*entities.Entity, ml *modellibrary.ModelLibrary) {
	for _, entity := range entitySlice {
		InitDeserializedEntity(entity, ml)
	}
}
