package libutils

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/lib/modelspec"
)

func ModelSpecVertsToVec3(vertices []modelspec.Vertex) []mgl64.Vec3 {
	var result []mgl64.Vec3

	for _, v := range vertices {
		result = append(result, Vec3F32ToF64(v.Position))
	}

	return result
}
