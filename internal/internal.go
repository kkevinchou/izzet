package lib

import (
	"errors"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func InitDepthCubeMap() (uint32, uint32) {
	var storedFBO int32
	gl.GetIntegerv(gl.FRAMEBUFFER_BINDING, &storedFBO)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, uint32(storedFBO))

	var depthCubeMapFBO uint32
	gl.GenFramebuffers(1, &depthCubeMapFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthCubeMapFBO)

	var depthCubeMapTexture uint32
	gl.GenTextures(1, &depthCubeMapTexture)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, depthCubeMapTexture)

	width, height := settings.DepthCubeMapWidth, settings.DepthCubeMapHeight
	for i := 0; i < 6; i++ {
		gl.TexImage2D(
			gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i),
			0,
			gl.DEPTH_COMPONENT,
			int32(width),
			int32(height),
			0,
			gl.DEPTH_COMPONENT,
			gl.FLOAT,
			nil,
		)
	}

	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, depthCubeMapTexture, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return depthCubeMapFBO, depthCubeMapTexture
}
