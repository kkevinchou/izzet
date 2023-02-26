package lib

import "github.com/go-gl/gl/v4.1-core/gl"

func InitDepthCubeMap() uint32 {

	var depthCubeMapFBO uint32
	gl.GenFramebuffers(1, &depthCubeMapFBO)

	var depthCubeMap uint32
	gl.GenTextures(1, &depthCubeMap)

	width, height := 1920, 1080
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, depthCubeMap)
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

	gl.BindFramebuffer(gl.FRAMEBUFFER, depthCubeMapFBO)
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, depthCubeMap, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	return depthCubeMap
}
