package apputils

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
)

// createUserSpaceTextureHandle creates a handle to a user space texture
// that the imgui renderer is able to render
func CreateUserSpaceTextureHandle(texture uint32) imgui.TextureID {
	handle := 1<<63 | uint64(texture)
	return imgui.TextureID(handle)
}

var ZeroVec = mgl64.Vec3{}
