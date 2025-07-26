package utils

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/modelspec"
)

func ModelSpecVertsToVec3(vertices []modelspec.Vertex) []mgl64.Vec3 {
	var result []mgl64.Vec3

	for _, v := range vertices {
		result = append(result, Vec3F32ToF64(v.Position))
	}

	return result
}

func PPrintVec(v mgl64.Vec3) string {
	return fmt.Sprintf("Vec3[%.1f, %.1f, %.1f]", v[0], v[1], v[2])
}

func PPrintQuatAsVec(q mgl64.Quat) string {
	return PPrintVec(q.Rotate(mgl64.Vec3{0, 0, -1}))
}

func PPrintVecList(vectors []mgl64.Vec3) string {
	var result string

	for _, v := range vectors {
		result += ", " + PPrintVec(v)
	}
	return "[ " + result[2:] + " ]"
}
