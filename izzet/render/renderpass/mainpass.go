package renderpass

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/gizmo"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
	"github.com/kkevinchou/izzet/izzet/render/rutils"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/shaders"
)

var (
	spatialPartitionLineCache [][2]mgl64.Vec3
)

type MainRenderPass struct {
	app renderiface.App
	sm  *shaders.ShaderManager

	// circle textures
	redCircleFB         uint32
	redCircleTexture    uint32
	greenCircleFB       uint32
	greenCircleTexture  uint32
	blueCircleFB        uint32
	blueCircleTexture   uint32
	yellowCircleFB      uint32
	yellowCircleTexture uint32
}

func NewMainPass(app renderiface.App, sm *shaders.ShaderManager) *MainRenderPass {
	return &MainRenderPass{app: app, sm: sm}
}

func (p *MainRenderPass) Init(width, height int, ctx *context.RenderPassContext) {
	fbo, textures := initFrameBuffer(width, height, []int32{rendersettings.InternalTextureColorFormatRGB, gl.R32UI}, []uint32{rendersettings.RenderFormatRGB, gl.RED_INTEGER}, []uint32{gl.FLOAT, gl.UNSIGNED_BYTE}, true)
	ctx.MainFBO = fbo
	ctx.MainTexture = textures[0]
	ctx.MainColorPickingTexture = textures[1]

	// init textures
	p.redCircleFB, p.redCircleTexture = createCircleTexture(1024, 1024)
	p.redCircleFB, p.redCircleTexture = createCircleTexture(1024, 1024)
	p.greenCircleFB, p.greenCircleTexture = createCircleTexture(1024, 1024)
	p.blueCircleFB, p.blueCircleTexture = createCircleTexture(1024, 1024)
	p.yellowCircleFB, p.yellowCircleTexture = createCircleTexture(1024, 1024)
	p.initializeCircleTextures()
}

func (p *MainRenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.MainFBO)
	textures := createAndBindTextures(width, height, []int32{rendersettings.InternalTextureColorFormatRGB, gl.R32UI}, []uint32{rendersettings.RenderFormatRGB, gl.RED_INTEGER}, []uint32{gl.FLOAT, gl.UNSIGNED_BYTE})
	ctx.MainTexture = textures[0]
}

func (p *MainRenderPass) Render(
	ctx context.RenderContext,
	rctx *context.RenderPassContext,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	lightViewerContext context.ViewerContext,
) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, rctx.MainFBO)
	gl.Viewport(0, 0, int32(ctx.Width()), int32(ctx.Height()))
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	mr := p.app.MetricsRegistry()

	// skybox
	p.drawSkybox(ctx, viewerContext)

	// models
	shader := p.sm.GetShaderProgram("modelpbr")
	p.renderModels(shader, viewerContext, lightContext, ctx, rctx, ctx.RenderableEntities)

	// colliders
	if p.app.RuntimeConfig().ShowColliders {
		p.drawColliders(viewerContext, ctx.RenderableEntities)
	}

	// non entities
	p.drawNonEntity(viewerContext, ctx)

	// annotations
	p.drawAnnotations(viewerContext, lightContext, ctx)

	// gizmos
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	start := time.Now()
	p.renderGizmos(viewerContext, ctx)
	mr.Inc("render_gizmos", float64(time.Since(start).Milliseconds()))
}

func (p *MainRenderPass) drawColliders(
	viewerContext context.ViewerContext,
	ents []*entities.Entity,
) {
	shader := p.sm.GetShaderProgram("flat")
	shader.Use()

	for _, entity := range ents {
		if entity == nil || entity.MeshComponent == nil || entity.Collider == nil {
			continue
		}

		if entity.MeshComponent.InvisibleToPlayerOwner && p.app.GetPlayerEntity().GetID() == entity.GetID() {
			continue
		}

		modelMatrix := entities.WorldTransform(entity)

		if entity.Collider.SimplifiedTriMeshCollider != nil {
			var lines [][2]mgl64.Vec3
			for _, triangles := range entity.Collider.SimplifiedTriMeshCollider.Triangles {
				lines = append(lines, [2]mgl64.Vec3{
					triangles.Points[0],
					triangles.Points[1],
				})
				lines = append(lines, [2]mgl64.Vec3{
					triangles.Points[1],
					triangles.Points[2],
				})
				lines = append(lines, [2]mgl64.Vec3{
					triangles.Points[2],
					triangles.Points[0],
				})
			}

			if len(lines) > 0 {
				scale := entity.Scale()
				shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
				shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
				shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
				rutils.DrawLineGroup(fmt.Sprintf("pogchamp_%d", len(lines)), shader, lines, 1/(scale.X()+scale.Y()+scale.Z())/3/8, mgl64.Vec3{1, 0, 0})
			}

			var pointLines [][2]mgl64.Vec3
			for _, p := range entity.Collider.SimplifiedTriMeshCollider.DebugPoints {
				// 0 length lines
				pointLines = append(pointLines, [2]mgl64.Vec3{p, p.Add(mgl64.Vec3{0.1, 0.1, 0.1})})
			}
			if len(pointLines) > 0 {
				rutils.DrawLineGroup(fmt.Sprintf("pogchamp_points_%d", len(pointLines)), shader, pointLines, 0.05, mgl64.Vec3{0, 0, 1})
			}
		}

		if entity.Collider.CapsuleCollider != nil {
			capsuleCollider := entity.Collider.CapsuleCollider

			top := capsuleCollider.Top
			bottom := capsuleCollider.Bottom
			radius := capsuleCollider.Radius

			var numCircleSegments int = 8
			var lines [][2]mgl64.Vec3

			// -x +x vertical lines
			lines = append(lines, [2]mgl64.Vec3{top.Add(mgl64.Vec3{-radius, 0, 0}), bottom.Add(mgl64.Vec3{-radius, 0, 0})})
			lines = append(lines, [2]mgl64.Vec3{bottom.Add(mgl64.Vec3{radius, 0, 0}), top.Add(mgl64.Vec3{radius, 0, 0})})

			// -z +z vertical lines
			lines = append(lines, [2]mgl64.Vec3{top.Add(mgl64.Vec3{0, 0, -radius}), bottom.Add(mgl64.Vec3{0, 0, -radius})})
			lines = append(lines, [2]mgl64.Vec3{bottom.Add(mgl64.Vec3{0, 0, radius}), top.Add(mgl64.Vec3{0, 0, radius})})

			radiansPerSegment := 2 * math.Pi / float64(numCircleSegments)

			// top and bottom xz plane rings
			for i := 0; i < numCircleSegments; i++ {
				x1 := math.Cos(float64(i)*radiansPerSegment) * radius
				z1 := math.Sin(float64(i)*radiansPerSegment) * radius

				x2 := math.Cos(float64((i+1)%numCircleSegments)*radiansPerSegment) * radius
				z2 := math.Sin(float64((i+1)%numCircleSegments)*radiansPerSegment) * radius

				lines = append(lines, [2]mgl64.Vec3{top.Add(mgl64.Vec3{x1, 0, -z1}), top.Add(mgl64.Vec3{x2, 0, -z2})})
				lines = append(lines, [2]mgl64.Vec3{bottom.Add(mgl64.Vec3{x1, 0, -z1}), bottom.Add(mgl64.Vec3{x2, 0, -z2})})
			}

			radiansPerSegment = math.Pi / float64(numCircleSegments)

			// top and bottom xy plane rings
			for i := 0; i < numCircleSegments; i++ {
				x1 := math.Cos(float64(i)*radiansPerSegment) * radius
				y1 := math.Sin(float64(i)*radiansPerSegment) * radius

				x2 := math.Cos(float64(float64(i+1)*radiansPerSegment)) * radius
				y2 := math.Sin(float64(float64(i+1)*radiansPerSegment)) * radius

				lines = append(lines, [2]mgl64.Vec3{top.Add(mgl64.Vec3{x1, y1, 0}), top.Add(mgl64.Vec3{x2, y2, 0})})
				lines = append(lines, [2]mgl64.Vec3{bottom.Add(mgl64.Vec3{x1, -y1, 0}), bottom.Add(mgl64.Vec3{x2, -y2, 0})})
			}

			// top and bottom yz plane rings
			for i := 0; i < numCircleSegments; i++ {
				z1 := math.Cos(float64(i)*radiansPerSegment) * radius
				y1 := math.Sin(float64(i)*radiansPerSegment) * radius

				z2 := math.Cos(float64(float64(i+1)*radiansPerSegment)) * radius
				y2 := math.Sin(float64(float64(i+1)*radiansPerSegment)) * radius

				lines = append(lines, [2]mgl64.Vec3{top.Add(mgl64.Vec3{0, y1, z1}), top.Add(mgl64.Vec3{0, y2, z2})})
				lines = append(lines, [2]mgl64.Vec3{bottom.Add(mgl64.Vec3{0, -y1, z1}), bottom.Add(mgl64.Vec3{0, -y2, z2})})
			}

			shader := p.sm.GetShaderProgram("flat")
			color := mgl64.Vec3{255.0 / 255, 147.0 / 255, 12.0 / 255}
			shader.Use()
			position := entity.Position()
			scale := entity.Scale()
			modelMat := mgl64.Translate3D(position.X(), position.Y(), position.Z()).Mul4(mgl64.Scale3D(scale.X(), scale.Y(), scale.Z()))
			shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMat))
			shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
			shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

			rutils.DrawLineGroup(fmt.Sprintf("%d_capsule_collider", entity.ID), shader, lines, 1/(scale.X()+scale.Y()+scale.Z())/3/8, color)
		}
	}
}

func (p *MainRenderPass) drawNonEntity(
	viewerContext context.ViewerContext,
	renderContext context.RenderContext,
) {
	// render non-models
	for _, entity := range p.app.World().Entities() {
		if entity.MeshComponent == nil {
			modelMatrix := entities.WorldTransform(entity)

			if len(entity.ShapeData) > 0 {
				shader := p.sm.GetShaderProgram("flat")
				shader.Use()

				shader.SetUniformUInt("entityID", uint32(entity.ID))
				shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix))
				shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
				shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
			}

			if entity.ImageInfo != nil {
				textureName := strings.Split(entity.ImageInfo.ImageName, ".")[0]
				texture := p.app.AssetManager().GetTexture(textureName)
				if texture != nil {
					if entity.Billboard && p.app.AppMode() == types.AppModeEditor {
						shader := p.sm.GetShaderProgram("world_space_quad")
						shader.Use()

						position := entity.Position()
						modelMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())
						scale := entity.ImageInfo.Scale
						modelMatrix = modelMatrix.Mul4(mgl64.Scale3D(scale, scale, scale))

						shader.SetUniformUInt("entityID", uint32(entity.ID))
						shader.SetUniformMat4("model", utils.Mat4F64ToF32(modelMatrix.Mul4(p.app.GetEditorCameraRotation().Mat4())))
						shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
						shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

						rutils.DrawBillboardTexture(texture.ID, 1)
					}
				} else {
					fmt.Println("couldn't find texture", "light")
				}
			}
			particles := entity.Particles
			if particles != nil {
				texture := p.app.AssetManager().GetTexture("light").ID
				for _, particle := range particles.GetActiveParticles() {
					particleModelMatrix := mgl32.Translate3D(float32(particle.Position.X()), float32(particle.Position.Y()), float32(particle.Position.Z()))
					rutils.DrawTexturedQuad(&viewerContext, p.sm, texture, float32(renderContext.AspectRatio()), &particleModelMatrix, true, nil)
				}
			}
		} else if entity.CharacterControllerComponent != nil {
			v := mgl64.Vec3{}
			if entity.CharacterControllerComponent.WebVector != v {
				// r.drawAABB(
				// 	viewerContext,
				// 	shaderManager.GetShaderProgram("flat"),
				// 	mgl64.Vec3{.2, 0, .7},
				// 	entity.BoundingBox(),
				// 	0.5,
				// )

				forwardVector := viewerContext.Rotation.Rotate(mgl64.Vec3{0, 0, -1})
				upVector := viewerContext.Rotation.Rotate(mgl64.Vec3{0, 1, 0})
				// there's probably away to get the right vector directly rather than going crossing the up vector :D
				rightVector := forwardVector.Cross(upVector)

				start := entity.Position().Add(rightVector.Mul(1)).Add(mgl64.Vec3{0, 2, 0})
				lines := [][2]mgl64.Vec3{
					{start, entity.Position().Add(entity.CharacterControllerComponent.WebVector)},
				}

				shader := p.sm.GetShaderProgram("flat")
				shader.Use()
				shader.SetUniformMat4("model", mgl32.Translate3D(float32(entity.Position().X()), float32(entity.Position().Y()), float32(entity.Position().Z())))
				shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
				shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

				rutils.DrawLineGroup(fmt.Sprintf("web_%d", len(lines)), shader, lines, 0.05, mgl64.Vec3{1, 0, 0})
			}
		}
	}
}

func (p *MainRenderPass) renderModels(shader *shaders.ShaderProgram,
	viewerContext context.ViewerContext,
	lightContext context.LightContext,
	renderContext context.RenderContext,
	renderPassContext *context.RenderPassContext,
	renderableEntities []*entities.Entity,
) {
	shader.Use()

	if p.app.RuntimeConfig().FogEnabled {
		shader.SetUniformInt("fog", 1)
	} else {
		shader.SetUniformInt("fog", 0)
	}

	var fog int32 = 0
	if p.app.RuntimeConfig().FogDensity != 0 {
		fog = 1
	}
	shader.SetUniformInt("fog", fog)
	shader.SetUniformInt("fogDensity", p.app.RuntimeConfig().FogDensity)

	shader.SetUniformInt("width", int32(renderContext.Width()))
	shader.SetUniformInt("height", int32(renderContext.Height()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformFloat("shadowDistance", renderContext.ShadowDistance)
	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	shader.SetUniformFloat("ambientFactor", p.app.RuntimeConfig().AmbientFactor)
	shader.SetUniformInt("shadowMap", 31)
	shader.SetUniformInt("depthCubeMap", 30)
	shader.SetUniformInt("cameraDepthMap", 29)
	shader.SetUniformInt("ambientOcclusion", 28)
	if p.app.RuntimeConfig().EnableSSAO {
		shader.SetUniformInt("enableAmbientOcclusion", 1)
	} else {
		shader.SetUniformInt("enableAmbientOcclusion", 0)
	}

	shader.SetUniformFloat("near", p.app.RuntimeConfig().Near)
	shader.SetUniformFloat("far", p.app.RuntimeConfig().Far)
	shader.SetUniformFloat("bias", p.app.RuntimeConfig().PointLightBias)
	if len(lightContext.PointLights) > 0 {
		shader.SetUniformFloat("far_plane", lightContext.PointLights[0].LightInfo.Range)
	}
	shader.SetUniformInt("hasColorOverride", 0)

	setupLightingUniforms(shader, lightContext.Lights)

	gl.ActiveTexture(gl.TEXTURE28)
	gl.BindTexture(gl.TEXTURE_2D, renderPassContext.SSAOBlurTexture)

	gl.ActiveTexture(gl.TEXTURE29)
	gl.BindTexture(gl.TEXTURE_2D, renderPassContext.CameraDepthTexture)

	gl.ActiveTexture(gl.TEXTURE30)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, renderPassContext.PointLightTexture)

	gl.ActiveTexture(gl.TEXTURE31)
	gl.BindTexture(gl.TEXTURE_2D, renderPassContext.ShadowMapTexture)

	var entityCount int
	for _, entity := range renderableEntities {
		if entity == nil || entity.MeshComponent == nil || !entity.MeshComponent.Visible {
			continue
		}

		if p.app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 && entity.Static {
			continue
		}

		if entity.MeshComponent.InvisibleToPlayerOwner && p.app.GetPlayerEntity().GetID() == entity.GetID() {
			continue
		}

		entityCount++

		shader.SetUniformUInt("entityID", uint32(entity.ID))

		p.drawModel(
			shader,
			entity,
		)
	}

	p.app.MetricsRegistry().Inc("draw_entity_count", float64(entityCount))

	if p.app.RuntimeConfig().BatchRenderingEnabled && len(renderContext.BatchRenders) > 0 {
		shader.SetUniformInt("hasColorOverride", 0)
		shader = p.sm.GetShaderProgram("batch")
		shader.Use()

		if p.app.RuntimeConfig().FogEnabled {
			shader.SetUniformInt("fog", 1)
		} else {
			shader.SetUniformInt("fog", 0)
		}

		var fog int32 = 0
		if p.app.RuntimeConfig().FogDensity != 0 {
			fog = 1
		}
		shader.SetUniformInt("fog", fog)
		shader.SetUniformInt("fogDensity", p.app.RuntimeConfig().FogDensity)

		shader.SetUniformInt("width", int32(renderContext.Width()))
		shader.SetUniformInt("height", int32(renderContext.Height()))
		shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
		shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
		shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
		shader.SetUniformFloat("shadowDistance", renderContext.ShadowDistance)
		shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
		shader.SetUniformFloat("ambientFactor", p.app.RuntimeConfig().AmbientFactor)
		shader.SetUniformInt("shadowMap", 31)
		shader.SetUniformInt("depthCubeMap", 30)
		shader.SetUniformInt("cameraDepthMap", 29)
		shader.SetUniformInt("ambientOcclusion", 28)
		if p.app.RuntimeConfig().EnableSSAO {
			shader.SetUniformInt("enableAmbientOcclusion", 1)
		} else {
			shader.SetUniformInt("enableAmbientOcclusion", 0)
		}

		shader.SetUniformFloat("near", p.app.RuntimeConfig().Near)
		shader.SetUniformFloat("far", p.app.RuntimeConfig().Far)
		shader.SetUniformFloat("bias", p.app.RuntimeConfig().PointLightBias)
		if len(lightContext.PointLights) > 0 {
			shader.SetUniformFloat("far_plane", lightContext.PointLights[0].LightInfo.Range)
		}
		shader.SetUniformInt("hasColorOverride", 0)

		setupLightingUniforms(shader, lightContext.Lights)

		gl.ActiveTexture(gl.TEXTURE28)
		gl.BindTexture(gl.TEXTURE_2D, renderPassContext.SSAOBlurTexture)

		gl.ActiveTexture(gl.TEXTURE29)
		gl.BindTexture(gl.TEXTURE_2D, renderPassContext.CameraDepthTexture)

		gl.ActiveTexture(gl.TEXTURE30)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, renderPassContext.PointLightTexture)

		gl.ActiveTexture(gl.TEXTURE31)
		gl.BindTexture(gl.TEXTURE_2D, renderPassContext.ShadowMapTexture)

		drawBatches(p.app, renderContext, shader)
		p.app.MetricsRegistry().Inc("draw_entity_count", 1)
	}
}

func (p *MainRenderPass) drawModel(
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
	primitives := p.app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
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
		primitiveMaterial := p.app.AssetManager().GetMaterial(materialHandle).Material
		material := primitiveMaterial.PBRMaterial.PBRMetallicRoughness

		if material.BaseColorTextureName != "" {
			shader.SetUniformInt("colorTextureCoordIndex", int32(material.BaseColorTextureCoordsIndex))
			shader.SetUniformInt("hasPBRBaseColorTexture", 1)

			textureName := primitiveMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName
			gl.ActiveTexture(gl.TEXTURE0)
			var textureID uint32
			texture := p.app.AssetManager().GetTexture(textureName)
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

func (p *MainRenderPass) drawSkybox(renderContext context.RenderContext, viewerContext context.ViewerContext) {
	if skyboxVAO == nil {
		var vbo, vao uint32
		apputils.GenBuffers(1, &vbo)
		gl.GenVertexArrays(1, &vao)

		gl.BindVertexArray(vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(skyboxVertices)*4, gl.Ptr(skyboxVertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
		gl.EnableVertexAttribArray(0)
		skyboxVAO = &vao
	}

	gl.DepthFunc(gl.LEQUAL)

	gl.BindVertexArray(*skyboxVAO)

	shader := p.sm.GetShaderProgram("skybox")
	shader.Use()
	var fog int32 = 0
	if p.app.RuntimeConfig().FogDensity != 0 {
		fog = 1
	}
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrixWithoutTranslation))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformInt("fog", fog)
	shader.SetUniformInt("fogDensity", p.app.RuntimeConfig().FogDensity)
	shader.SetUniformFloat("far", p.app.RuntimeConfig().Far)
	shader.SetUniformVec3("skyboxTopColor", p.app.RuntimeConfig().SkyboxTopColor)
	shader.SetUniformVec3("skyboxBottomColor", p.app.RuntimeConfig().SkyboxBottomColor)
	shader.SetUniformFloat("skyboxMixValue", p.app.RuntimeConfig().SkyboxMixValue)
	rutils.IztDrawArrays(0, 36)
	gl.DepthFunc(gl.LESS)
}

var skyboxVAO *uint32
var skyboxVertices = []float32{
	// Front face
	-1.0, 1.0, -1.0,
	-1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	1.0, 1.0, -1.0,
	-1.0, 1.0, -1.0,

	// Left face
	-1.0, -1.0, 1.0,
	-1.0, -1.0, -1.0,
	-1.0, 1.0, -1.0,
	-1.0, 1.0, -1.0,
	-1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0,

	// Right face
	1.0, -1.0, -1.0,
	1.0, -1.0, 1.0,
	1.0, 1.0, 1.0,
	1.0, 1.0, 1.0,
	1.0, 1.0, -1.0,
	1.0, -1.0, -1.0,

	// Back face
	-1.0, -1.0, 1.0,
	-1.0, 1.0, 1.0,
	1.0, 1.0, 1.0,
	1.0, 1.0, 1.0,
	1.0, -1.0, 1.0,
	-1.0, -1.0, 1.0,

	// Top face
	-1.0, 1.0, -1.0,
	1.0, 1.0, -1.0,
	1.0, 1.0, 1.0,
	1.0, 1.0, 1.0,
	-1.0, 1.0, 1.0,
	-1.0, 1.0, -1.0,

	// Bottom face
	-1.0, -1.0, -1.0,
	-1.0, -1.0, 1.0,
	1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	-1.0, -1.0, 1.0,
	1.0, -1.0, 1.0,
}

// i considered using uniform blocks but the memory layout management seems like a huge pain
// https://stackoverflow.com/questions/38172696/should-i-ever-use-a-vec3-inside-of-a-uniform-buffer-or-shader-storage-buffer-o
func setupLightingUniforms(shader *shaders.ShaderProgram, lights []*entities.Entity) {
	if len(lights) > settings.MaxLightCount {
		panic(fmt.Sprintf("light count of %d exceeds max %d", len(lights), settings.MaxLightCount))
	}

	shader.SetUniformInt("lightCount", int32(len(lights)))
	for i, light := range lights {
		lightInfo := light.LightInfo

		diffuse := lightInfo.IntensifiedDiffuse()

		shader.SetUniformInt(fmt.Sprintf("lights[%d].type", i), int32(lightInfo.Type))
		shader.SetUniformVec3(fmt.Sprintf("lights[%d].dir", i), lightInfo.Direction3F)
		shader.SetUniformVec3(fmt.Sprintf("lights[%d].diffuse", i), diffuse)
		shader.SetUniformVec3(fmt.Sprintf("lights[%d].position", i), utils.Vec3F64ToF32(light.Position()))
		shader.SetUniformFloat(fmt.Sprintf("lights[%d].range", i), lightInfo.Range)
	}
}

func (p *MainRenderPass) drawAnnotations(viewerContext context.ViewerContext, lightContext context.LightContext, renderContext context.RenderContext) {
	if p.app.RuntimeConfig().ShowSelectionBoundingBox {
		entity := p.app.SelectedEntity()
		if entity != nil {
			// draw bounding box
			if entity.HasBoundingBox() {
				shader := p.sm.GetShaderProgram("flat")
				shader.Use()
				rutils.DrawAABB(
					shader,
					viewerContext,
					mgl64.Vec3{.2, 0, .7},
					entity.BoundingBox(),
					0.1,
				)
			}
		}
	}

	if p.app.AppMode() == types.AppModeEditor {
		for _, entity := range p.app.World().Entities() {
			lightInfo := entity.LightInfo
			if lightInfo != nil {
				if lightInfo.Type == 0 {

					direction3F := lightInfo.Direction3F
					dir := mgl64.Vec3{float64(direction3F[0]), float64(direction3F[1]), float64(direction3F[2])}.Mul(5)
					// directional light arrow
					lines := [][2]mgl64.Vec3{
						[2]mgl64.Vec3{
							entity.Position(),
							entity.Position().Add(dir),
						},
					}

					shader := p.sm.GetShaderProgram("flat")
					color := mgl64.Vec3{252.0 / 255, 241.0 / 255, 33.0 / 255}
					shader.Use()
					shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
					shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
					shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

					rutils.DrawLineGroup(fmt.Sprintf("%d_%v_%v", entity.ID, entity.Position(), dir), shader, lines, 0.05, color)
				}
			}
		}
	}

	if p.app.RuntimeConfig().RenderSpatialPartition {
		p.drawSpatialPartition(viewerContext, mgl64.Vec3{0, 1, 0}, p.app.World().SpatialPartition(), 0.1)
	}

	nm := p.app.NavMesh()
	if nm != nil {
		shader := p.sm.GetShaderProgram("navmesh")
		shader.Use()
		shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
		shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
		shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

		setupLightingUniforms(shader, lightContext.Lights)
		shader.SetUniformInt("width", int32(renderContext.Width()))
		shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
		shader.SetUniformFloat("shadowDistance", renderContext.ShadowDistance)
		shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
		shader.SetUniformFloat("ambientFactor", p.app.RuntimeConfig().AmbientFactor)
		shader.SetUniformInt("shadowMap", 31)
		shader.SetUniformInt("depthCubeMap", 30)
		shader.SetUniformInt("cameraDepthMap", 29)
		shader.SetUniformFloat("near", p.app.RuntimeConfig().Near)
		shader.SetUniformFloat("far", p.app.RuntimeConfig().Far)
		shader.SetUniformFloat("bias", p.app.RuntimeConfig().PointLightBias)
		if len(lightContext.PointLights) > 0 {
			shader.SetUniformFloat("far_plane", lightContext.PointLights[0].LightInfo.Range)
		}
		shader.SetUniformVec3("albedo", mgl32.Vec3{1, 0, 0})

		shader.SetUniformFloat("roughness", .8)
		shader.SetUniformFloat("metallic", 0)

		p.drawNavmesh(p.sm, viewerContext, nm)

		// draw bounding box
		volume := nm.Volume
		rutils.DrawAABB(
			p.sm.GetShaderProgram("flat"),
			viewerContext,
			mgl64.Vec3{155.0 / 99, 180.0 / 255, 45.0 / 255},
			volume,
			0.5,
		)

		if len(nm.DebugLines) > 0 {
			shader := p.sm.GetShaderProgram("flat")
			// color := mgl64.Vec3{252.0 / 255, 241.0 / 255, 33.0 / 255}
			color := mgl64.Vec3{1, 0, 0}
			shader.Use()
			shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
			shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
			shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
			rutils.DrawLineGroup(fmt.Sprintf("navmesh_debuglines_%d", nm.InvalidatedTimestamp), shader, nm.DebugLines, 0.05, color)
		}

		nm.Invalidated = false
	}
}

func (p *MainRenderPass) drawSpatialPartition(viewerContext context.ViewerContext, color mgl64.Vec3, spatialPartition *spatialpartition.SpatialPartition, thickness float64) {
	var allLines [][2]mgl64.Vec3

	if len(spatialPartitionLineCache) == 0 {
		d := spatialPartition.PartitionDimension * spatialPartition.PartitionCount
		var baseHorizontalLines [][]mgl64.Vec3

		// lines along z axis
		for i := 0; i < spatialPartition.PartitionCount+1; i++ {
			baseHorizontalLines = append(baseHorizontalLines,
				[]mgl64.Vec3{{float64(-d/2 + i*spatialPartition.PartitionDimension), float64(-d / 2), float64(-d / 2)}, {float64(-d/2 + i*spatialPartition.PartitionDimension), float64(-d / 2), float64(d / 2)}},
			)
		}

		// // lines along x axis
		for i := 0; i < spatialPartition.PartitionCount+1; i++ {
			baseHorizontalLines = append(baseHorizontalLines,
				[]mgl64.Vec3{{float64(-d / 2), float64(-d / 2), float64(-d/2 + i*spatialPartition.PartitionDimension)}, {float64(d / 2), float64(-d / 2), float64(-d/2 + i*spatialPartition.PartitionDimension)}},
			)
		}

		for i := 0; i < spatialPartition.PartitionCount+1; i++ {
			for _, b := range baseHorizontalLines {
				allLines = append(allLines,
					[2]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, float64(i * spatialPartition.PartitionDimension), 0}), b[1].Add(mgl64.Vec3{0, float64(i * spatialPartition.PartitionDimension), 0})},
				)
			}
		}

		var baseVerticalLines [][]mgl64.Vec3

		for i := 0; i < spatialPartition.PartitionCount+1; i++ {
			baseVerticalLines = append(baseVerticalLines,
				[]mgl64.Vec3{{float64(-d/2 + i*spatialPartition.PartitionDimension), float64(-d / 2), float64(-d / 2)}, {float64(-d/2 + i*spatialPartition.PartitionDimension), float64(d / 2), float64(-d / 2)}},
			)
		}

		for i := 0; i < spatialPartition.PartitionCount+1; i++ {
			for _, b := range baseVerticalLines {
				allLines = append(allLines,
					[2]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, 0, float64(i * spatialPartition.PartitionDimension)}), b[1].Add(mgl64.Vec3{0, 0, float64(i * spatialPartition.PartitionDimension)})},
				)
			}
		}
		spatialPartitionLineCache = allLines
	}
	allLines = spatialPartitionLineCache

	shader := p.sm.GetShaderProgram("flat")
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	rutils.DrawLineGroup("spatial_partition", shader, allLines, thickness, color)
}

func (p *MainRenderPass) renderGizmos(viewerContext context.ViewerContext, renderContext context.RenderContext) {
	if p.app.SelectedEntity() == nil {
		return
	}

	entity := p.app.World().GetEntityByID(p.app.SelectedEntity().ID)
	position := entity.Position()

	if gizmo.CurrentGizmoMode == gizmo.GizmoModeTranslation {
		p.drawTranslationGizmo(&viewerContext, p.sm.GetShaderProgram("flat"), position)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeRotation {
		p.drawCircleGizmo(&viewerContext, position, renderContext)
	} else if gizmo.CurrentGizmoMode == gizmo.GizmoModeScale {
		p.drawScaleGizmo(&viewerContext, p.sm.GetShaderProgram("flat"), position)
	}
}

func createCircleTexture(width, height int) (uint32, uint32) {
	fbo, textures := initFrameBuffer(width, height, []int32{rendersettings.InternalTextureColorFormatRGBA}, []uint32{rendersettings.RenderFormatRGBA}, []uint32{gl.UNSIGNED_BYTE}, true)
	return fbo, textures[0]
}

// setup reusale circle textures
func (p *MainRenderPass) initializeCircleTextures() {
	gl.Viewport(0, 0, 1024, 1024)
	shader := p.sm.GetShaderProgram("unit_circle")
	shader.Use()

	gl.BindFramebuffer(gl.FRAMEBUFFER, p.redCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{1, 0, 0, 1})
	drawCircle()

	gl.BindFramebuffer(gl.FRAMEBUFFER, p.greenCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{0, 1, 0, 1})
	drawCircle()

	gl.BindFramebuffer(gl.FRAMEBUFFER, p.blueCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{0, 0, 1, 1})
	drawCircle()

	gl.BindFramebuffer(gl.FRAMEBUFFER, p.yellowCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{1, 1, 0, 1})
	drawCircle()
}

func drawCircle() {
	var vertices []float32 = []float32{
		-1, -1, 0,
		1, -1, 0,
		1, 1, 0,
		1, 1, 0,
		-1, 1, 0,
		-1, -1, 0,
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(vao)

	rutils.IztDrawArrays(0, 6)
}
