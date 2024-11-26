package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/utils"
)

func (r *RenderSystem) SetupBatchedStaticRendering() {
	var meshHandles []types.MeshHandle

	var modelMatrices []mgl32.Mat4
	var ids []uint32
	for _, entity := range r.app.World().Entities() {
		if entity.MeshComponent == nil || !entity.Static {
			continue
		}

		meshHandle := entity.MeshComponent.MeshHandle
		meshHandles = append(meshHandles, meshHandle)

		modelMatrix := entities.WorldTransform(entity)
		modelMat := utils.Mat4F64ToF32(modelMatrix).Mul4(utils.Mat4F64ToF32(entity.MeshComponent.Transform))

		modelMatrices = append(modelMatrices, modelMat)
		ids = append(ids, uint32(entity.GetID()))
	}
	r.batchRenders = r.app.AssetManager().SetupBatchedStaticRendering(meshHandles, modelMatrices, ids)
}
