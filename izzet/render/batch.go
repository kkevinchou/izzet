package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/types"
)

func (r *RenderSystem) SetupBatchedStaticRendering() {
	var meshHandles []types.MeshHandle

	var modelMatrices []mgl32.Mat4
	var ids []uint32
	for _, e := range r.app.World().Entities() {
		if !entity.BatchRenderable(e) {
			continue
		}

		meshHandle := e.MeshComponent.MeshHandle
		meshHandles = append(meshHandles, meshHandle)

		modelMatrix := entity.WorldTransform(e)
		modelMat := utils.Mat4F64ToF32(modelMatrix).Mul4(utils.Mat4F64ToF32(e.MeshComponent.Transform))

		modelMatrices = append(modelMatrices, modelMat)
		ids = append(ids, uint32(e.GetID()))
	}
	r.batchRenders = r.app.AssetManager().SetupBatchedStaticRendering(meshHandles, modelMatrices, ids)
}
