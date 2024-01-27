package apputils

import (
	"math"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func NameFromAssetFilePath(assetFilePath string) string {
	return strings.Split(filepath.Base(assetFilePath), ".")[0]
}

func GenBuffers(n int32, buffer *uint32) {
	mr := globals.GetClientMetricsRegistry()
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

func CalculateFooterSize(uiEnabled bool) float32 {
	if !uiEnabled {
		return 0
	}
	return 34
}

func PathToProjectFile(projectName string) string {
	return filepath.Join(settings.ProjectsDirectory, projectName, "main_project.izt")
}
