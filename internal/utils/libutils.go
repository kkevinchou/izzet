package utils

import (
	"fmt"
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/modelspec"
)

type JointWeight struct {
	JointID int
	Weight  float32
}

type ByWeights []JointWeight

func (s ByWeights) Len() int           { return len(s) }
func (s ByWeights) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByWeights) Less(i, j int) bool { return s[i].Weight < s[j].Weight }

func FillWeights(jointIDs []int, weights []float32, maxAnimationJointWeights int) ([]int, []float32) {
	var resultJointIDs []int
	var resultWeights []float32

	if len(jointIDs) <= maxAnimationJointWeights {
		resultJointIDs = append(resultJointIDs, jointIDs...)
		resultWeights = append(resultWeights, weights...)
		for i := 0; i < maxAnimationJointWeights-len(jointIDs); i++ {
			resultJointIDs = append(resultJointIDs, 0)
			resultWeights = append(resultWeights, 0)
		}
	} else {
		jointWeights := make([]JointWeight, 0, len(jointIDs))
		for i := range jointIDs {
			jointWeights = append(jointWeights, JointWeight{JointID: jointIDs[i], Weight: weights[i]})
		}
		sort.Sort(sort.Reverse(ByWeights(jointWeights)))

		jointWeights = jointWeights[:maxAnimationJointWeights]
		NormalizeWeights(jointWeights)
		for _, jointWeight := range jointWeights {
			resultJointIDs = append(resultJointIDs, jointWeight.JointID)
			resultWeights = append(resultWeights, jointWeight.Weight)
		}
	}

	return resultJointIDs, resultWeights
}

func NormalizeWeights(jointWeights []JointWeight) {
	var totalWeight float32
	for _, jointWeight := range jointWeights {
		totalWeight += jointWeight.Weight
	}

	for i := range jointWeights {
		jointWeights[i].Weight /= totalWeight
	}
}

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
