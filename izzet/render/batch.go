package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/render/batch"
)

func (s *RenderSystem) SetupBatchedStaticRendering() {
	var meshHandles []assets.MeshHandle
	var modelMatrices []mgl32.Mat4
	var materials [][]assets.MaterialID
	var ids []uint32

	for _, e := range s.app.World().Entities() {
		if !entity.BatchRenderable(e) {
			continue
		}

		materials = append(materials, entity.GetPrimitiveMaterialIDs(s.app.AssetManager(), e))

		meshHandle := e.MeshComponent.MeshHandle
		meshHandles = append(meshHandles, meshHandle)

		modelMatrix := entity.WorldTransform(e)
		modelMat := utils.Mat4F64ToF32(modelMatrix).Mul4(utils.Mat4F64ToF32(e.MeshComponent.Transform))

		modelMatrices = append(modelMatrices, modelMat)
		ids = append(ids, uint32(e.GetID()))
	}
	s.batchRenders = batch.SetupBatchedStaticRendering(s.app.AssetManager(), meshHandles, modelMatrices, ids, materials)
}
