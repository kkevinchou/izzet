package renderpass

import (
	"fmt"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/shaders"
)

const (
	gPassInternalFormat int32  = gl.RGB32F
	gPassFormat         uint32 = gl.RGB
)

type GBufferPass struct {
	app    renderiface.App
	shader *shaders.ShaderProgram
}

func (p *GBufferPass) Name() string { return "G-Buffer" }

func (p *GBufferPass) Init(app renderiface.App, width, height int, sm *shaders.ShaderManager, ctx *context.RenderPassContext) error {
	p.app = app

	// create FBO + 3 render targets
	geometryTextureFn := textureFn(width, height,
		[]int32{gPassInternalFormat, gPassInternalFormat, gPassInternalFormat},
		[]uint32{gPassFormat, gPassFormat, gPassFormat},
		[]uint32{gl.FLOAT, gl.FLOAT, gl.FLOAT},
	)
	geometryFBO, textures := initFrameBuffer(geometryTextureFn)
	ctx.GeometryFBO = geometryFBO
	ctx.GPositionTexture, ctx.GNormalTexture, ctx.GColorTexture = textures[0], textures[1], textures[2]
	p.shader = sm.GetShaderProgram("gpass")
	return nil
}

func (p *GBufferPass) Resize(width, height int, ctx *context.RenderPassContext) {
	// re-alloc textures on size change
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.GeometryFBO)

	_, _, textures := textureFn(width, height,
		[]int32{gPassInternalFormat, gPassInternalFormat, gPassInternalFormat},
		[]uint32{gPassFormat, gPassFormat, gPassFormat},
		[]uint32{gl.FLOAT, gl.FLOAT, gl.FLOAT},
	)()

	ctx.GPositionTexture, ctx.GNormalTexture, ctx.GColorTexture = textures[0], textures[1], textures[2]
}

func (p *GBufferPass) Render(ctx context.RenderContext, rctx *context.RenderPassContext, viewerContext context.ViewerContext, ents []*entities.Entity) {
	mr := p.app.MetricsRegistry()
	start := time.Now()

	// bind, clear, draw
	gl.BindFramebuffer(gl.FRAMEBUFFER, rctx.GeometryFBO)
	gl.Viewport(0, 0, int32(ctx.Width()), int32(ctx.Height()))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	p.shader.Use()

	p.shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	p.shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	for _, entity := range ents {
		if entity == nil || entity.MeshComponent == nil || !entity.MeshComponent.Visible {
			continue
		}

		var animationPlayer *animation.AnimationPlayer
		if entity.Animation != nil {
			animationPlayer = entity.Animation.AnimationPlayer
		}

		if animationPlayer != nil && animationPlayer.CurrentAnimation() != "" {
			p.shader.SetUniformInt("isAnimated", 1)
			animationTransforms := animationPlayer.AnimationTransforms()

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

		primitives := p.app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
		for _, prim := range primitives {
			modelMatrix := entities.WorldTransform(entity)
			var modelMat mgl32.Mat4

			// apply smooth blending between mispredicted position and actual real position
			if entity.RenderBlend != nil && entity.RenderBlend.Active {
				deltaMs := time.Since(entity.RenderBlend.StartTime).Milliseconds()
				t := apputils.RenderBlendMath(deltaMs)

				blendedPosition := entity.Position().Sub(entity.RenderBlend.BlendStartPosition).Mul(t).Add(entity.RenderBlend.BlendStartPosition)

				translationMatrix := mgl64.Translate3D(blendedPosition[0], blendedPosition[1], blendedPosition[2])
				rotationMatrix := entity.GetLocalRotation().Mat4()
				scale := entity.Scale()
				scaleMatrix := mgl64.Scale3D(scale.X(), scale.Y(), scale.Z())
				modelMatrix = translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)

				if deltaMs >= int64(settings.RenderBlendDurationMilliseconds) {
					entity.RenderBlend.Active = false
				}
			}

			modelMat = utils.Mat4F64ToF32(modelMatrix).Mul4(utils.Mat4F64ToF32(entity.MeshComponent.Transform))

			p.shader.SetUniformMat4("model", modelMat)
			gl.BindVertexArray(prim.VAO)
			iztDrawElements(p.app, int32(len(prim.Primitive.VertexIndices)))
		}
	}

	mr.Inc("render_gpass", float64(time.Since(start).Milliseconds()))
}
