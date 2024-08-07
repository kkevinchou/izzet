package render

import (
	"errors"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type ShadowMap struct {
	depthMapFBO    uint32
	depthTexture   uint32
	width          int
	height         int
	shadowDistance float64
}

func (s *ShadowMap) DepthMapFBO() uint32 {
	return s.depthMapFBO
}

func (s *ShadowMap) Prepare() {
	gl.CullFace(gl.FRONT)
	gl.Viewport(0, 0, int32(s.height), int32(s.height))
	gl.BindFramebuffer(gl.FRAMEBUFFER, s.depthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
}

func NewShadowMap(width int, height int, far float64) (*ShadowMap, error) {
	var storedFBO int32
	gl.GetIntegerv(gl.FRAMEBUFFER_BINDING, &storedFBO)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, uint32(storedFBO))

	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT,
		int32(width), int32(height), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		return nil, errors.New("failed to initialize shadow map frame buffer - in the past this was due to an overly large shadow map dimension configuration")
	}

	return &ShadowMap{
		depthMapFBO:    depthMapFBO,
		depthTexture:   texture,
		width:          width,
		height:         height,
		shadowDistance: far * settings.ShadowMapDistanceFactor,
	}, nil
}

func (s *ShadowMap) DepthTexture() uint32 {
	return s.depthTexture
}

func (s *ShadowMap) ShadowDistance() float64 {
	return s.shadowDistance
}

// returns the orthographic projection matrix for the directional light as well as the "position" of the light
func ComputeDirectionalLightProps(lightRotationMatrix mgl64.Mat4, frustumPoints []mgl64.Vec3, shadowMapZOffset float32) (mgl64.Vec3, mgl64.Mat4) {
	var lightSpacePoints []mgl64.Vec3
	invLightRotationMatrix := lightRotationMatrix.Inv()

	for _, point := range frustumPoints {
		lightSpacePoint := invLightRotationMatrix.Mul4x1(point.Vec4(1)).Vec3()
		lightSpacePoints = append(lightSpacePoints, lightSpacePoint)
	}

	var minX, maxX, minY, maxY, minZ, maxZ float64

	minX = lightSpacePoints[0].X()
	maxX = lightSpacePoints[0].X()
	minY = lightSpacePoints[0].Y()
	maxY = lightSpacePoints[0].Y()
	minZ = lightSpacePoints[0].Z()
	maxZ = lightSpacePoints[0].Z()

	for _, point := range lightSpacePoints {
		if point.X() < minX {
			minX = point.X()
		}
		if point.X() > maxX {
			maxX = point.X()
		}
		if point.Y() < minY {
			minY = point.Y()
		}
		if point.Y() > maxY {
			maxY = point.Y()
		}
		if point.Z() < minZ {
			minZ = point.Z()
		}
		if point.Z() > maxZ {
			maxZ = point.Z()
		}
	}
	maxZ += float64(shadowMapZOffset)

	halfX := (maxX - minX) / 2
	halfY := (maxY - minY) / 2
	halfZ := (maxZ - minZ) / 2
	position := mgl64.Vec3{minX + halfX, minY + halfY, maxZ}
	position = lightRotationMatrix.Mul4x1(position.Vec4(1)).Vec3() // bring position back into world space
	orthoProjMatrix := mgl64.Ortho(-halfX, halfX, -halfY, halfY, 0, halfZ*2)
	return position, orthoProjMatrix
}
