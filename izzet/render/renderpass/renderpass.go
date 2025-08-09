package renderpass

import (
	"errors"
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/shaders"
)

type TextureFn func() (int, int, []uint32)

// RenderPass is a single step in the frame‐render pipeline.
type RenderPass interface {
	// Init is called once at startup (or when switching pipelines)
	Init(width, height int, ctx *context.RenderPassContext)

	// Resize is called whenever the viewport changes size
	Resize(width, height int, ctx *context.RenderPassContext)

	// Render executes the pass. It may read from
	// previous-output textures and write into its own FBO.
	Render(ctx context.RenderContext, rctx *context.RenderPassContext, viewerContext context.ViewerContext)
}

func initFrameBuffer(tf TextureFn) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	width, height, textures := tf()
	var drawBuffers []uint32

	textureCount := len(textures)
	for i := 0; i < textureCount; i++ {
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(textureCount), &drawBuffers[0])

	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(width), int32(height))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, textures
}

func initDepthMapFrameBuffer(texture uint32) uint32 {
	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("failed to initialize shadow map frame buffer - in the past this was due to an overly large shadow map dimension configuration")
	}

	return depthMapFBO
}

func createDepthTexture(width, height int) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT,
		int32(width), int32(height), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return texture
}

func initFrameBufferNoDepth(tf TextureFn) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	_, _, textures := tf()
	var drawBuffers []uint32

	textureCount := len(textures)
	for i := 0; i < textureCount; i++ {
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(textureCount), &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, textures
}

func textureFn(width int, height int, internalFormat []int32, format []uint32, xtype []uint32) func() (int, int, []uint32) {
	return func() (int, int, []uint32) {
		count := len(internalFormat)
		var textures []uint32
		for i := 0; i < count; i++ {
			texture := createTexture(width, height, internalFormat[i], format[i], xtype[i], gl.LINEAR)
			attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachment, gl.TEXTURE_2D, texture, 0)

			textures = append(textures, texture)
		}
		return width, height, textures
	}
}

func createTexture(width, height int, internalFormat int32, format uint32, xtype uint32, filtering int32) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat,
		int32(width), int32(height), 0, format, xtype, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return texture
}

func iztDrawElements(app renderiface.App, count int32) {
	app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	app.RuntimeConfig().DrawCount += 1
	gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, nil)
}

func iztDrawArrays(app renderiface.App, first, count int32) {
	app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	app.RuntimeConfig().DrawCount += 1
	gl.DrawArrays(gl.TRIANGLES, first, count)
}

func createNDCQuadVAO() uint32 {
	vertices := []float32{
		-1, -1, 0.0, 0.0,
		1, -1, 1.0, 0.0,
		1, 1, 1.0, 1.0,
		1, 1, 1.0, 1.0,
		-1, 1, 0.0, 1.0,
		-1, -1, 0.0, 0.0,
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	return vao
}

func renderGeometryWithoutColor(
	viewerContext context.ViewerContext,
	shader *shaders.ShaderProgram,
	app renderiface.App,
	renderContext context.RenderContext,
	renderableEntities []*entities.Entity,
) {
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	for _, entity := range renderableEntities {
		if entity.MeshComponent == nil {
			continue
		}

		if app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 && entity.Static {
			continue
		}

		if entity.Animation != nil && entity.Animation.AnimationPlayer.CurrentAnimation() != "" {
			shader.SetUniformInt("isAnimated", 1)
			animationTransforms := entity.Animation.AnimationPlayer.AnimationTransforms()
			// if animationTransforms is nil, the shader will execute reading into invalid memory
			// so, we need to explicitly guard for this
			if animationTransforms == nil {
				panic("animationTransforms not found")
			}
			for i := 0; i < len(animationTransforms); i++ {
				shader.SetUniformMat4(fmt.Sprintf("jointTransforms[%d]", i), animationTransforms[i])
			}
		} else {
			shader.SetUniformInt("isAnimated", 0)
		}

		modelMatrix := entities.WorldTransform(entity)
		m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix)

		primitives := app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
		for _, p := range primitives {
			shader.SetUniformMat4("model", m32ModelMatrix.Mul4(utils.Mat4F64ToF32(entity.MeshComponent.Transform)))

			gl.BindVertexArray(p.GeometryVAO)
			iztDrawElements(app, int32(len(p.Primitive.VertexIndices)))
		}
	}

	if app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 {
		drawBatches(app, renderContext, shader)
		app.MetricsRegistry().Inc("draw_entity_count", 1)
	}
}

func drawBatches(
	app renderiface.App,
	renderContext context.RenderContext,
	shader *shaders.ShaderProgram,
) {
	shader.SetUniformInt("isAnimated", 0)
	shader.SetUniformMat4("model", mgl32.Scale3D(1, 1, 1))

	for _, batch := range renderContext.BatchRenders {
		primitiveMaterial := app.AssetManager().GetMaterial(batch.MaterialHandle).Material

		material := primitiveMaterial.PBRMaterial.PBRMetallicRoughness
		shader.SetUniformInt("colorTextureCoordIndex", int32(material.BaseColorTextureCoordsIndex))

		shader.SetUniformInt("hasPBRBaseColorTexture", 1)
		shader.SetUniformVec3("albedo", material.BaseColorFactor.Vec3())
		shader.SetUniformFloat("roughness", material.RoughnessFactor)
		shader.SetUniformFloat("metallic", material.MetalicFactor)

		if material.BaseColorTextureName != "" {
			shader.SetUniformInt("colorTextureCoordIndex", int32(material.BaseColorTextureCoordsIndex))
			shader.SetUniformInt("hasPBRBaseColorTexture", 1)

			textureName := material.BaseColorTextureName
			gl.ActiveTexture(gl.TEXTURE0)
			var textureID uint32
			texture := app.AssetManager().GetTexture(textureName)
			textureID = texture.ID
			gl.BindTexture(gl.TEXTURE_2D, textureID)
		} else {
			shader.SetUniformInt("hasPBRBaseColorTexture", 0)
		}

		shader.SetUniformVec3("albedo", material.BaseColorFactor.Vec3())
		shader.SetUniformFloat("roughness", material.RoughnessFactor)
		shader.SetUniformFloat("metallic", material.MetalicFactor)

		gl.BindVertexArray(batch.VAO)
		iztDrawElements(app, batch.VertexCount)
	}
}
