package izzet

import (
	"fmt"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/model"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/utils"
)

type ViewerContext struct {
	Position    mgl64.Vec3
	Orientation mgl64.Quat

	InverseViewMatrix mgl64.Mat4
	ProjectionMatrix  mgl64.Mat4
}

type LightContext struct {
	DirectionalLightDir mgl64.Vec3
	LightSpaceMatrix    mgl64.Mat4
}

var (
	defaultTexture string = "color_grid"
	aspectRatio           = float64(settings.Width) / float64(settings.Height)

	// shadow map properties
	shadowmapZOffset     float64 = 400
	fovx                 float64 = 105
	fovy                         = mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovx)/2)/aspectRatio))
	near                 float64 = 1
	far                  float64 = 3000
	shadowDistanceFactor float64 = .4 // proportion of view fustrum to include in shadow cuboid
)

var update float64

func toRadians(degrees float64) float64 {
	return degrees / 180 * math.Pi
}

func (g *Izzet) Render(delta time.Duration) {
	// configure camera viewer context
	position := mgl64.Vec3{0, 0, 0}
	orientation := mgl64.QuatIdent()

	update += float64(delta.Milliseconds())
	orientation = mgl64.QuatRotate(toRadians(float64(int(update/1000*360)%360)), mgl64.Vec3{0, 1, 0})

	viewerViewMatrix := orientation.Mat4()
	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())

	cameraViewerContext := ViewerContext{
		Position:    position,
		Orientation: orientation,

		InverseViewMatrix: viewTranslationMatrix.Mul4(viewerViewMatrix).Inv(),
		ProjectionMatrix:  mgl64.Perspective(mgl64.DegToRad(fovy), aspectRatio, near, far),
	}

	// configure light viewer context
	modelSpaceFrustumPoints := CalculateFrustumPoints(position, orientation, near, far, fovy, aspectRatio, shadowDistanceFactor)

	lightOrientation := utils.Vec3ToQuat(mgl64.Vec3{-1, -1, -1})
	lightPosition, lightProjectionMatrix := ComputeDirectionalLightProps(lightOrientation.Mat4(), modelSpaceFrustumPoints, shadowmapZOffset)
	lightViewMatrix := mgl64.Translate3D(lightPosition.X(), lightPosition.Y(), lightPosition.Z()).Mul4(lightOrientation.Mat4()).Inv()

	lightViewerContext := ViewerContext{
		Position:          lightPosition,
		Orientation:       lightOrientation,
		InverseViewMatrix: lightViewMatrix,
		ProjectionMatrix:  lightProjectionMatrix,
	}

	lightContext := LightContext{
		DirectionalLightDir: lightOrientation.Rotate(mgl64.Vec3{0, 0, -1}),
		// this should be the inverse of the transforms applied to the viewer context
		// if the viewer moves along -y, the universe moves along +y
		LightSpaceMatrix: lightProjectionMatrix.Mul4(lightViewMatrix),
	}

	_ = cameraViewerContext
	_ = lightViewerContext
	_ = lightContext
	// s.renderToDepthMap(lightViewerContext, lightContext)
	// g.renderToDisplay(cameraViewerContext, lightContext)
	g.renderToDisplay2()
	// s.renderImgui()

	g.window.GLSwap()
}

func (g *Izzet) renderToDisplay(viewerContext ViewerContext, lightContext LightContext) {
	defer resetGLRenderSettings()

	gl.Viewport(0, 0, int32(settings.Width), int32(settings.Height))
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	g.renderScene(viewerContext, lightContext, false)
}

var count int

func (g *Izzet) renderToDisplay2() {
	count += 1
	defer resetGLRenderSettings()

	gl.Viewport(0, 0, int32(settings.Width), int32(settings.Height))
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	if count%2 == 0 {
		g.shaderManager.GetShaderProgram("ndc").Use()
		gl.BindVertexArray(g.vao1)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
	} else {
		g.shaderManager.GetShaderProgram("ndc").Use()
		gl.BindVertexArray(g.vao2)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		g.window.GLSwap()
	}
}

// renderScene renders a scene from the perspective of a viewer
func (g *Izzet) renderScene(viewerContext ViewerContext, lightContext LightContext, shadowPass bool) {
	shaderManager := g.shaderManager

	// meshModelMatrix := createModelMatrix(
	// 	mgl64.Scale3D(100, 100, 100),
	// 	mgl64.QuatIdent().Mat4(),
	// 	mgl64.Ident4(),
	// )

	// shader := "model_static"
	// if componentContainer.AnimationComponent != nil {
	// 	shader = "modelpbr"
	// }

	// drawModel(
	// 	viewerContext,
	// 	lightContext,
	// 	g.shadowMap,
	// 	shaderManager.GetShaderProgram("model_static"),
	// 	g.assetManager,
	// 	g.model,
	// 	nil,
	// 	meshModelMatrix,
	// )

	drawTris(viewerContext, shaderManager.GetShaderProgram("flat"), []mgl64.Vec3{{0, 0, 0}, {100, 0, 100}, {100, 100, 100}}, mgl64.Vec3{1, 0, 0})
	drawTris(viewerContext, shaderManager.GetShaderProgram("flat"), []mgl64.Vec3{{0, 0, 0}, {100, 100, 100}, {100, 0, 100}}, mgl64.Vec3{1, 0, 0})

	drawTris(viewerContext, shaderManager.GetShaderProgram("flat"), []mgl64.Vec3{{0, 0, 0}, {100, 0, -100}, {100, 100, -100}}, mgl64.Vec3{1, 0, 0})
	drawTris(viewerContext, shaderManager.GetShaderProgram("flat"), []mgl64.Vec3{{0, 0, 0}, {100, 100, -100}, {100, 0, -100}}, mgl64.Vec3{1, 0, 0})
}

func createModelMatrix(scaleMatrix, rotationMatrix, translationMatrix mgl64.Mat4) mgl64.Mat4 {
	return translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)
}

// drawTris draws a list of triangles in winding order. each triangle is defined with 3 consecutive points
func drawTris(viewerContext ViewerContext, shader *shaders.ShaderProgram, points []mgl64.Vec3, color mgl64.Vec3) {
	var vertices []float32
	for _, point := range points {
		vertices = append(vertices, float32(point.X()), float32(point.Y()), float32(point.Z()))
	}

	var vbo, vao uint32
	gl.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(vao)
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformFloat("alpha", float32(1))
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
}

func drawModel(viewerContext ViewerContext,
	lightContext LightContext,
	shadowMap *ShadowMap,
	shader *shaders.ShaderProgram,
	assetManager *assets.AssetManager,
	model *model.Model,
	animationPlayer *animation.AnimationPlayer,
	modelMatrix mgl64.Mat4,
) {
	// TOOD(kevin): i hate this... Ideally we incorporate the model.RootTransforms to the vertex positions
	// and the animation poses so that we don't have to multiple this matrix every frame.
	m32ModelMatrix := utils.Mat4F64ToF32(modelMatrix).Mul4(model.RootTransforms())
	_, r, _ := utils.Decompose(m32ModelMatrix)

	shader.Use()
	shader.SetUniformMat4("model", m32ModelMatrix)
	shader.SetUniformMat4("modelRotationMatrix", r.Mat4())
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformVec3("viewPos", utils.Vec3F64ToF32(viewerContext.Position))
	shader.SetUniformFloat("shadowDistance", float32(shadowMap.ShadowDistance()))
	shader.SetUniformVec3("directionalLightDir", utils.Vec3F64ToF32(lightContext.DirectionalLightDir))
	shader.SetUniformMat4("lightSpaceMatrix", utils.Mat4F64ToF32(lightContext.LightSpaceMatrix))
	shader.SetUniformInt("shadowMap", 31)

	if animationPlayer != nil {
		animationTransforms := animationPlayer.AnimationTransforms()
		// if animationTransforms is nil, the shader will execute reading into invalid memory
		// so, we need to explicitly guard for this
		if animationTransforms == nil {
			panic("animationTransforms not found")
		}
		for i := 0; i < len(animationTransforms); i++ {
			shader.SetUniformMat4(fmt.Sprintf("jointTransforms[%d]", i), animationTransforms[i])
		}
	}

	gl.ActiveTexture(gl.TEXTURE31)
	gl.BindTexture(gl.TEXTURE_2D, shadowMap.DepthTexture())

	for _, meshChunk := range model.MeshChunks() {
		if pbr := meshChunk.PBRMaterial(); pbr != nil {
			shader.SetUniformInt("hasPBRMaterial", 1)
			shader.SetUniformVec4("pbrBaseColorFactor", pbr.PBRMetallicRoughness.BaseColorFactor)

			if pbr.PBRMetallicRoughness.BaseColorTextureIndex != nil {
				shader.SetUniformInt("hasPBRBaseColorTexture", 1)
			} else {
				shader.SetUniformInt("hasPBRBaseColorTexture", 0)
			}

			shader.SetUniformVec3("albedo", pbr.PBRMetallicRoughness.BaseColorFactor.Vec3())
			shader.SetUniformFloat("metallic", pbr.PBRMetallicRoughness.MetalicFactor)
			shader.SetUniformFloat("roughness", pbr.PBRMetallicRoughness.RoughnessFactor)
			shader.SetUniformFloat("ao", 1.0)
		} else {
			shader.SetUniformInt("hasPBRMaterial", 0)
		}

		gl.ActiveTexture(gl.TEXTURE0)
		var textureID uint32
		if meshChunk.TextureID() != nil {
			textureID = *meshChunk.TextureID()
		} else {
			texture := assetManager.GetTexture(defaultTexture)
			textureID = texture.ID
		}
		gl.BindTexture(gl.TEXTURE_2D, textureID)

		gl.BindVertexArray(meshChunk.VAO())
		gl.DrawElements(gl.TRIANGLES, int32(meshChunk.VertexCount()), gl.UNSIGNED_INT, nil)
	}
}
