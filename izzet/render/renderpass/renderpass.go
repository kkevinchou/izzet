package renderpass

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/globals"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/shaders"
)

// RenderPass is a single step in the frame‐render pipeline.
type RenderPass interface {
	// Init is called once at startup (or when switching pipelines)
	Init(width, height int, ctx *context.RenderPassContext)

	// Resize is called whenever the viewport changes size
	Resize(width, height int, ctx *context.RenderPassContext)

	// Render executes the pass. It may read from
	// previous-output textures and write into its own FBO.
	Render(
		renderContext context.RenderContext,
		renderPassContext *context.RenderPassContext,
		viewerContext context.ViewerContext,
		lightContext context.LightContext,
		lightViewerContext context.ViewerContext,
	)
}

func initFrameBuffer(width int, height int, internalFormat []int32, format []uint32, xtype []uint32, includeDepth bool, singleSample bool) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var drawBuffers []uint32

	textures := createAndBindTextures(width, height, internalFormat, format, xtype, singleSample)
	textureCount := len(textures)
	for i := 0; i < textureCount; i++ {
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(textureCount), &drawBuffers[0])

	if includeDepth {
		var rbo uint32
		gl.GenRenderbuffers(1, &rbo)
		gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
		if singleSample {
			gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(width), int32(height))
		} else {
			gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, 4, gl.DEPTH_COMPONENT, int32(width), int32(height))
		}
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)
	}

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, textures
}

func initDepthOnlyFrameBuffer(width, height int) (uint32, uint32) {
	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)

	texture := createDepthTexture(width, height)

	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("failed to initialize shadow map frame buffer - in the past this was due to an overly large shadow map dimension configuration")
	}

	return depthMapFBO, texture
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

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texture, 0)

	return texture
}

func createAndBindTextures(width int, height int, internalFormat []int32, format []uint32, xtype []uint32, singleSample bool) []uint32 {
	count := len(internalFormat)
	var textures []uint32
	for i := 0; i < count; i++ {
		texture := createTexture(width, height, internalFormat[i], format[i], xtype[i], gl.LINEAR, singleSample)
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		if singleSample {
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachment, gl.TEXTURE_2D, texture, 0)
		} else {
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachment, gl.TEXTURE_2D_MULTISAMPLE, texture, 0)
		}

		textures = append(textures, texture)
	}
	return textures
}

func createTexture(width, height int, internalFormat int32, format uint32, xtype uint32, filtering int32, singleSample bool) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)

	if singleSample {
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat,
			int32(width), int32(height), 0, format, xtype, nil)
	} else {
		gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, texture)
		gl.TexImage2DMultisample(gl.TEXTURE_2D_MULTISAMPLE, 4, format,
			int32(width), int32(height), true)
	}

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return texture
}

func renderGeometryWithoutColor(
	app renderiface.App,
	shader *shaders.ShaderProgram,
	ents []*entities.Entity,
	viewerContext context.ViewerContext,
	renderContext context.RenderContext,
) {
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	for _, entity := range ents {
		if entity.MeshComponent == nil {
			continue
		}

		if app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 && entities.BatchRenderable(entity) {
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
			rutils.IztDrawElements(int32(len(p.Primitive.VertexIndices)))
		}
	}

	if app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 {
		drawBatches(app, renderContext, shader)
		globals.ClientRegistry().Inc("draw_entity_count", 1)
	}
}

func drawBatches(
	app renderiface.App,
	renderContext context.RenderContext,
	shader *shaders.ShaderProgram,
) {
	shader.SetUniformInt("isAnimated", 0)
	shader.SetUniformInt("alphaMode", int32(modelspec.AlphaModeOpaque))
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
		rutils.IztDrawElements(batch.VertexCount)
	}
}
func worldToNDCPosition(viewerContext context.ViewerContext, worldPosition mgl64.Vec3) (mgl64.Vec2, bool) {
	screenPos := viewerContext.ProjectionMatrix.Mul4(viewerContext.InverseViewMatrix).Mul4x1(worldPosition.Vec4(1))
	behind := screenPos.Z() < 0
	screenPos = screenPos.Mul(1 / screenPos.W())
	return screenPos.Vec2(), behind
}

func drawModels(
	app renderiface.App,
	renderShader *shaders.ShaderProgram,
	batchShader *shaders.ShaderProgram,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	ents []*entities.Entity,
) {
	gl.ActiveTexture(gl.TEXTURE28)
	gl.BindTexture(gl.TEXTURE_2D, renderPassContext.SSAOBlurTexture)

	gl.ActiveTexture(gl.TEXTURE29)
	gl.BindTexture(gl.TEXTURE_2D, renderPassContext.CameraDepthTexture)

	gl.ActiveTexture(gl.TEXTURE30)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, renderPassContext.PointLightTexture)

	gl.ActiveTexture(gl.TEXTURE31)
	gl.BindTexture(gl.TEXTURE_2D, renderPassContext.ShadowMapTexture)

	renderShader.Use()
	preModelRenderShaderSetup(app, renderShader, renderContext, viewerContext, lightContext)

	var drawCount int
	for _, entity := range ents {
		if app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 && entities.BatchRenderable(entity) {
			continue
		}
		renderShader.SetUniformUInt("entityID", uint32(entity.ID))
		drawModel(
			app,
			renderShader,
			entity,
		)
		drawCount++
	}
	globals.ClientRegistry().Inc("draw_entity_count", float64(drawCount))

	if app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 {
		batchShader.Use()
		preModelRenderShaderSetup(app, batchShader, renderContext, viewerContext, lightContext)
		drawBatches(app, renderContext, batchShader)
		globals.ClientRegistry().Inc("draw_entity_count", 1)
	}
}

func preModelRenderShaderSetup(app renderiface.App, shader *shaders.ShaderProgram, renderContext context.RenderContext, viewerContext context.ViewerContext, lightContext context.LightContext) {
	shader.SetUniformInt("fogDensity", app.RuntimeConfig().FogDensity)

	shader.SetUniformInt("width", int32(renderContext.Width()))
	shader.SetUniformInt("height", int32(renderContext.Height()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformFloat("shadowDistance", renderContext.ShadowDistance)
	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	shader.SetUniformFloat("ambientFactor", app.RuntimeConfig().AmbientFactor)
	shader.SetUniformFloat("specularFactor", app.RuntimeConfig().SpecularFactor)
	shader.SetUniformInt("shadowMap", 31)
	shader.SetUniformInt("depthCubeMap", 30)
	shader.SetUniformInt("cameraDepthMap", 29)
	shader.SetUniformInt("ambientOcclusion", 28)
	if app.RuntimeConfig().EnableSSAO {
		shader.SetUniformInt("enableAmbientOcclusion", 1)
	} else {
		shader.SetUniformInt("enableAmbientOcclusion", 0)
	}

	shader.SetUniformFloat("near", app.RuntimeConfig().Near)
	shader.SetUniformFloat("far", app.RuntimeConfig().Far)
	shader.SetUniformFloat("pointLightBias", app.RuntimeConfig().PointLightBias)
	shader.SetUniformFloat("shadowMapMinBias", app.RuntimeConfig().ShadowMapMinBias/100000)
	shader.SetUniformFloat("shadowMapAngleBiasRate", app.RuntimeConfig().ShadowMapAngleBiasRate/100000)
	if len(lightContext.PointLights) > 0 {
		shader.SetUniformFloat("far_plane", lightContext.PointLights[0].LightInfo.Range)
	}

	setupLightingUniforms(shader, lightContext.Lights)
}

func drawModel(
	app renderiface.App,
	shader *shaders.ShaderProgram,
	entity *entities.Entity,
) {
	var animationPlayer *animation.AnimationPlayer
	if entity.Animation != nil {
		animationPlayer = entity.Animation.AnimationPlayer
	}

	if animationPlayer != nil && animationPlayer.CurrentAnimation() != "" {
		shader.SetUniformInt("isAnimated", 1)
		animationTransforms := animationPlayer.AnimationTransforms()

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

	// THE HOTTEST CODE PATH IN THE ENGINE
	primitives := app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
	if entity.MeshComponent.MeshHandle == assets.DefaultCubeHandle {
		shader.SetUniformInt("repeatTexture", 1)
	} else {
		shader.SetUniformInt("repeatTexture", 0)
	}
	for _, prim := range primitives {
		materialHandle := prim.MaterialHandle
		if entity.Material != nil {
			materialHandle = entity.Material.MaterialHandle
		}
		primitiveMaterial := app.AssetManager().GetMaterial(materialHandle).Material
		material := primitiveMaterial.PBRMaterial.PBRMetallicRoughness
		alphaMode := primitiveMaterial.PBRMaterial.AlphaMode

		if material.BaseColorTextureName != "" {
			shader.SetUniformInt("colorTextureCoordIndex", int32(material.BaseColorTextureCoordsIndex))
			shader.SetUniformInt("hasPBRBaseColorTexture", 1)

			textureName := primitiveMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName
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
		shader.SetUniformVec3("translation", utils.Vec3F64ToF32(entity.Position()))
		shader.SetUniformVec3("scale", utils.Vec3F64ToF32(entity.Scale()))
		shader.SetUniformInt("alphaMode", int32(alphaMode))

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

		shader.SetUniformMat4("model", modelMat)

		gl.BindVertexArray(prim.VAO)
		if modelMat.Det() < 0 {
			// from the gltf spec:
			// When a mesh primitive uses any triangle-based topology (i.e., triangles, triangle strip, or triangle fan),
			// the determinant of the node’s global transform defines the winding order of that primitive. If the determinant
			// is a positive value, the winding order triangle faces is counterclockwise; in the opposite case, the winding
			// order is clockwise.
			gl.FrontFace(gl.CW)
		}
		rutils.IztDrawElements(int32(len(prim.Primitive.VertexIndices)))
		if modelMat.Det() < 0 {
			gl.FrontFace(gl.CCW)
		}
	}
}
