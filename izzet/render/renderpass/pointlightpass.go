package renderpass

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/kitolib/shaders"
)

const (
	PointLightCubeMapWidth  float32 = 4096
	PointLightCubeMapHeight float32 = 4096
	PointLightCubeMapNear   float64 = 1
)

type PointLightRenderPass struct {
	app    renderiface.App
	shader *shaders.ShaderProgram
}

func NewPointLightPass(app renderiface.App, sm *shaders.ShaderManager) *PointLightRenderPass {
	return &PointLightRenderPass{app: app, shader: sm.GetShaderProgram("point_shadow")}
}

func (p *PointLightRenderPass) Init(_, _ int, ctx *context.RenderPassContext) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, texture)

	width, height := PointLightCubeMapWidth, PointLightCubeMapHeight
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

	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, texture, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	ctx.PointLightFBO = fbo
	ctx.PointLightTexture = texture
}

func (p *PointLightRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
}

func (p *PointLightRenderPass) Render(
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	start := time.Now()
	defer func() {
		globals.ClientRegistry().Inc("render_pointlight_pass", float64(time.Since(start).Milliseconds()))
	}()

	// we only support cube depth maps for one point light atm
	var pointLight *entity.Entity
	if len(lightContext.PointLights) == 0 {
		return
	}
	pointLight = lightContext.PointLights[0]

	gl.Viewport(0, 0, int32(PointLightCubeMapWidth), int32(PointLightCubeMapHeight))
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderPassContext.PointLightFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	position := pointLight.Position()
	shadowTransforms := computeCubeMapTransforms(position, PointLightCubeMapNear, float64(lightContext.PointLights[0].LightInfo.Range))

	p.shader.Use()
	for i, transform := range shadowTransforms {
		p.shader.SetUniformMat4(fmt.Sprintf("shadowMatrices[%d]", i), utils.Mat4F64ToF32(transform))
	}
	if len(lightContext.PointLights) > 0 {
		p.shader.SetUniformFloat("far_plane", lightContext.PointLights[0].LightInfo.Range)
	}
	p.shader.SetUniformVec3("lightPos", utils.Vec3F64ToF32(position))

	for _, e := range renderContext.RenderableEntities {
		if e == nil || e.MeshComponent == nil {
			continue
		}

		if p.app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 && entity.BatchRenderable(e) {
			continue
		}

		if e.Animation != nil && e.Animation.AnimationPlayer.CurrentAnimation() != "" {
			p.shader.SetUniformInt("isAnimated", 1)
			animationTransforms := e.Animation.AnimationPlayer.AnimationTransforms()
			// if animationTransforms is nil, the shader will execute reading into invalid memory
			// so, we need to explicitly guard for this
			if animationTransforms == nil {
				panic("animationTransforms not found")
			}
			for i := 0; i < len(animationTransforms); i++ {
				p.shader.SetUniformMat4(fmt.Sprintf("jointTransforms[%d]", i), animationTransforms[i])
			}
		} else {
			p.shader.SetUniformInt("isAnimated", 0)
		}

		modelMatrix := entity.WorldTransform(e)
		m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix)

		primitives := p.app.AssetManager().GetPrimitives(e.MeshComponent.MeshHandle)
		for _, primitive := range primitives {
			p.shader.SetUniformMat4("model", m32ModelMatrix.Mul4(utils.Mat4F64ToF32(e.MeshComponent.Transform)))

			gl.BindVertexArray(primitive.GeometryVAO)
			rutils.IztDrawElements(int32(len(primitive.Primitive.VertexIndices)))
		}
	}
	if p.app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 {
		drawBatches(p.app, renderContext, p.shader)
		globals.ClientRegistry().Inc("draw_entity_count", 1)
	}
}

func computeCubeMapTransforms(position mgl64.Vec3, near, far float64) []mgl64.Mat4 {
	projectionMatrix := mgl64.Perspective(
		mgl64.DegToRad(90),
		float64(PointLightCubeMapWidth)/float64(PointLightCubeMapHeight),
		near,
		far,
	)

	cubeMapTransforms := []mgl64.Mat4{
		projectionMatrix.Mul4( // right
			mgl64.LookAtV(
				position,
				position.Add(mgl64.Vec3{1, 0, 0}),
				mgl64.Vec3{0, -1, 0},
			),
		),
		projectionMatrix.Mul4( // left
			mgl64.LookAtV(
				position,
				position.Add(mgl64.Vec3{-1, 0, 0}),
				mgl64.Vec3{0, -1, 0},
			),
		),
		projectionMatrix.Mul4( // up
			mgl64.LookAtV(
				position,
				position.Add(mgl64.Vec3{0, 1, 0}),
				mgl64.Vec3{0, 0, 1},
			),
		),
		projectionMatrix.Mul4( // down
			mgl64.LookAtV(
				position,
				position.Add(mgl64.Vec3{0, -1, 0}),
				mgl64.Vec3{0, 0, -1},
			),
		),
		projectionMatrix.Mul4( // back
			mgl64.LookAtV(
				position,
				position.Add(mgl64.Vec3{0, 0, 1}),
				mgl64.Vec3{0, -1, 0},
			),
		),
		projectionMatrix.Mul4( // front
			mgl64.LookAtV(
				position,
				position.Add(mgl64.Vec3{0, 0, -1}),
				mgl64.Vec3{0, -1, 0},
			),
		),
	}
	return cubeMapTransforms
}
