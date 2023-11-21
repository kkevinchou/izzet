package apputils

import (
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/globals"
)

// createUserSpaceTextureHandle creates a handle to a user space texture
// that the imgui renderer is able to render
func CreateUserSpaceTextureHandle(texture uint32) imgui.TextureID {
	handle := 1<<63 | uint64(texture)
	return imgui.TextureID(handle)
}

func NameFromAssetFilePath(assetFilePath string) string {
	return strings.Split(filepath.Base(assetFilePath), ".")[0]
}

func GenBuffers(n int32, buffer *uint32) {
	mr := globals.GetClientMetricsRegistry()
	mr.Inc("gen_buffers", 1)
	gl.GenBuffers(n, buffer)
}

var ZeroVec = mgl64.Vec3{}
