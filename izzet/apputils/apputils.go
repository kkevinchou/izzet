package apputils

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/input"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func NameFromAssetFilePath(assetFilePath string) string {
	return strings.Split(filepath.Base(assetFilePath), ".")[0]
}

func GenBuffers(n int32, buffer *uint32) {
	mr := globals.ClientRegistry()
	mr.Inc("gen_buffers", 1)
	gl.GenBuffers(n, buffer)
}

var ZeroVec = mgl64.Vec3{}

func Vec3ApproxEqualThreshold(v1 mgl64.Vec3, v2 mgl64.Vec3, threshold float64) bool {
	return v1.ApproxFuncEqual(v2, func(a, b float64) bool {
		return math.Abs(a-b) < threshold
	})
}

func Vec4ApproxEqualThreshold(v1 mgl64.Vec4, v2 mgl64.Vec4, threshold float64) bool {
	return v1.ApproxFuncEqual(v2, func(a, b float64) bool {
		return math.Abs(a-b) < threshold
	})
}

func GetDrawerbarSize(uiEnabled bool) float32 {
	if !uiEnabled {
		return 0
	}
	return settings.DrawerbarSize
}

func PathToProjectFile(projectName string) string {
	return filepath.Join(settings.ProjectsDirectory, projectName, "main_project.izt")
}

var zeroVec mgl64.Vec3

func IsZeroVec(v mgl64.Vec3) bool {
	return v == zeroVec
}

func GetControlVector(keyboardInput input.KeyboardInput) mgl64.Vec3 {
	var controlVector mgl64.Vec3
	if _, ok := keyboardInput[input.KeyboardKeyW]; ok {
		controlVector[2]++
	}
	if _, ok := keyboardInput[input.KeyboardKeyS]; ok {
		controlVector[2]--
	}
	if _, ok := keyboardInput[input.KeyboardKeyA]; ok {
		controlVector[0]--
	}
	if _, ok := keyboardInput[input.KeyboardKeyD]; ok {
		controlVector[0]++
	}
	if _, ok := keyboardInput[input.KeyboardKeyLShift]; ok {
		controlVector[1]--
	}
	if _, ok := keyboardInput[input.KeyboardKeySpace]; ok {
		controlVector[1]++
	}

	return controlVector
}

func RenderBlendMath(deltaMs int64) float64 {
	var t float64 = float64(deltaMs) / settings.RenderBlendDurationMilliseconds
	if t >= 1 {
		t = 1
	} else {
		t = 1 - math.Pow(2, -10*t)
	}
	return t
}

func FormatVec(vec mgl64.Vec3) string {
	return fmt.Sprintf("{%.2f, %.2f, %.2f}", vec.X(), vec.Y(), vec.Z())
}

func ModelSpecVertsToVec3(vertices []modelspec.Vertex) []mgl64.Vec3 {
	var result []mgl64.Vec3

	for _, v := range vertices {
		result = append(result, utils.Vec3F32ToF64(v.Position))
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
