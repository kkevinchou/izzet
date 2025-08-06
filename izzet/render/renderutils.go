package render

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/animation"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/shaders"
)

var lineCache map[string][]mgl64.Vec3
var cubeCache map[string][]mgl64.Vec3
var triangleVAOCache map[string]TriangleVAO
var singleSidedQuadVAO uint32
var pickingBuffer []byte

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

var (
	spatialPartitionLineCache [][2]mgl64.Vec3
)

func init() {
	lineCache = map[string][]mgl64.Vec3{}
	cubeCache = map[string][]mgl64.Vec3{}
	triangleVAOCache = map[string]TriangleVAO{}
}

type TriangleVAO struct {
	VAO    uint32
	length int
}

type RenderData struct {
	Primitive   *modelspec.PrimitiveSpecification
	Transform   mgl32.Mat4
	VAO         uint32
	GeometryVAO uint32
}

type TextureFn func() (int, int, []uint32)

func genCacheKey(thickness, length float64) string {
	return fmt.Sprintf("%.3f_%.3f", thickness, length)
}

func (r *RenderSystem) generateTrisVAO(points []mgl64.Vec3) (uint32, int) {
	var vertices []float32
	for _, point := range points {
		vertices = append(vertices, float32(point.X()), float32(point.Y()), float32(point.Z()))
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	return vao, len(vertices)
}

func RGBtoHSV(rgb mgl32.Vec3) mgl32.Vec3 {
	// Normalize RGB values to be between 0 and 1
	r := rgb.X()
	g := rgb.Y()
	b := rgb.Z()

	// Determine maximum and minimum values among R, G, and B
	maxVal := float32(math.Max(math.Max(float64(r), float64(g)), float64(b)))
	minVal := float32(math.Min(math.Min(float64(r), float64(g)), float64(b)))

	// Calculate value (V) as maximum of R, G, and B
	v := maxVal

	// Calculate saturation (S)
	var s float32
	if maxVal == 0 {
		s = 0
	} else {
		s = (maxVal - minVal) / maxVal
	}

	// Calculate hue (H)
	var h float32
	if maxVal == minVal {
		h = 0
	} else if maxVal == r && g >= b {
		h = 60 * (g - b) / (maxVal - minVal)
	} else if maxVal == r && g < b {
		h = 60*(g-b)/(maxVal-minVal) + 360
	} else if maxVal == g {
		h = 60*(b-r)/(maxVal-minVal) + 120
	} else { // maxVal == B
		h = 60*(r-g)/(maxVal-minVal) + 240
	}

	// Return HSV values as an mgl32.Vec3
	return mgl32.Vec3{h, s, v}
}

func HSVtoRGB(hsv mgl32.Vec3) mgl32.Vec3 {
	// Extract H, S, and V values from input Vec3
	h := hsv.X()
	s := hsv.Y()
	v := hsv.Z()

	// Calculate chroma (C)
	c := v * s

	// Calculate h' (hPrime)
	hPrime := h / 60

	// Calculate x
	x := c * float32(1-math.Abs(float64(math.Mod(float64(hPrime), 2)-1)))

	// Calculate m
	m := v - c

	// Initialize RGB values to m
	r := m
	g := m
	b := m

	// Determine which sector of the color wheel h' falls in and set RGB values accordingly
	if hPrime < 1 {
		r += c
		g += x
	} else if hPrime < 2 {
		r += x
		g += c
	} else if hPrime < 3 {
		g += c
		b += x
	} else if hPrime < 4 {
		g += x
		b += c
	} else if hPrime < 5 {
		r += x
		b += c
	} else {
		r += c
		b += x
	}

	// Create and return RGB Vec3
	return mgl32.Vec3{r, g, b}
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

func (r *RenderSystem) drawBatches(
	shader *shaders.ShaderProgram,
) {
	shader.SetUniformInt("isAnimated", 0)
	shader.SetUniformMat4("model", mgl32.Scale3D(1, 1, 1))

	for _, batch := range r.batchRenders {
		primitiveMaterial := r.app.AssetManager().GetMaterial(batch.MaterialHandle).Material

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
			texture := r.app.AssetManager().GetTexture(textureName)
			textureID = texture.ID
			gl.BindTexture(gl.TEXTURE_2D, textureID)
		} else {
			shader.SetUniformInt("hasPBRBaseColorTexture", 0)
		}

		shader.SetUniformVec3("albedo", material.BaseColorFactor.Vec3())
		shader.SetUniformFloat("roughness", material.RoughnessFactor)
		shader.SetUniformFloat("metallic", material.MetalicFactor)

		gl.BindVertexArray(batch.VAO)
		r.iztDrawElements(batch.VertexCount)
	}
}

func (r *RenderSystem) drawSpherePBR(shader *shaders.ShaderProgram, materialHandle types.MaterialHandle, vao uint32, vertexCount int32) {
	shader.SetUniformInt("isAnimated", 0)
	shader.SetUniformMat4("model", mgl32.Scale3D(1, 1, 1))

	primitiveMaterial := r.app.AssetManager().GetMaterial(materialHandle).Material

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
		texture := r.app.AssetManager().GetTexture(textureName)
		textureID = texture.ID
		gl.BindTexture(gl.TEXTURE_2D, textureID)
	} else {
		shader.SetUniformInt("hasPBRBaseColorTexture", 0)
	}

	gl.BindVertexArray(vao)
	r.iztDrawElements(vertexCount)
}

func (r *RenderSystem) drawModel(
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
	primitives := r.app.AssetManager().GetPrimitives(entity.MeshComponent.MeshHandle)
	if entity.MeshComponent.MeshHandle == assets.DefaultCubeHandle {
		shader.SetUniformInt("repeatTexture", 1)
	} else {
		shader.SetUniformInt("repeatTexture", 0)
	}
	for _, p := range primitives {
		materialHandle := p.MaterialHandle
		if entity.Material != nil {
			materialHandle = entity.Material.MaterialHandle
		}
		primitiveMaterial := r.app.AssetManager().GetMaterial(materialHandle).Material
		material := primitiveMaterial.PBRMaterial.PBRMetallicRoughness

		if material.BaseColorTextureName != "" {
			shader.SetUniformInt("colorTextureCoordIndex", int32(material.BaseColorTextureCoordsIndex))
			shader.SetUniformInt("hasPBRBaseColorTexture", 1)

			textureName := primitiveMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName
			gl.ActiveTexture(gl.TEXTURE0)
			var textureID uint32
			texture := r.app.AssetManager().GetTexture(textureName)
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

		gl.BindVertexArray(p.VAO)
		if modelMat.Det() < 0 {
			// from the gltf spec:
			// When a mesh primitive uses any triangle-based topology (i.e., triangles, triangle strip, or triangle fan),
			// the determinant of the node’s global transform defines the winding order of that primitive. If the determinant
			// is a positive value, the winding order triangle faces is counterclockwise; in the opposite case, the winding
			// order is clockwise.
			gl.FrontFace(gl.CW)
		}
		r.iztDrawElements(int32(len(p.Primitive.VertexIndices)))
		if modelMat.Det() < 0 {
			gl.FrontFace(gl.CCW)
		}
	}
}

func (r *RenderSystem) drawLineGroup(name string, shader *shaders.ShaderProgram, lines [][2]mgl64.Vec3, thickness float64, color mgl64.Vec3) {
	var vao uint32
	var length int

	if _, ok := triangleVAOCache[name]; !ok {
		var points []mgl64.Vec3
		for _, line := range lines {
			start := line[0]
			end := line[1]
			length := end.Sub(start).Len()

			if length == 0 {
				cp := cubePoints(thickness)
				for _, p := range cp {
					points = append(points, p.Add(start))
				}
			} else {
				dir := end.Sub(start).Normalize()
				q := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)

				for _, dp := range linePoints(thickness, length) {
					newEnd := q.Rotate(dp).Add(start)
					points = append(points, newEnd)
				}
			}
		}
		vao, length = r.generateTrisVAO(points)
		item := TriangleVAO{VAO: vao, length: length}
		triangleVAOCache[name] = item
	}

	item := triangleVAOCache[name]
	vao = item.VAO
	length = item.length

	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	shader.SetUniformFloat("intensity", 1.0)
	gl.BindVertexArray(vao)
	r.iztDrawArrays(0, int32(length))
}

func cubePoints(thickness float64) []mgl64.Vec3 {
	cacheKey := genCacheKey(thickness, 0)
	if _, ok := cubeCache[cacheKey]; ok {
		return cubeCache[cacheKey]
	}

	var ht float64 = thickness / 2
	points := []mgl64.Vec3{
		// front
		{-ht, -ht, ht},
		{ht, -ht, ht},
		{ht, ht, ht},

		{ht, ht, ht},
		{-ht, ht, ht},
		{-ht, -ht, ht},

		// back
		{ht, ht, -ht},
		{ht, -ht, -ht},
		{-ht, -ht, -ht},

		{-ht, -ht, -ht},
		{-ht, ht, -ht},
		{ht, ht, -ht},

		// right
		{ht, -ht, ht},
		{ht, -ht, -ht},
		{ht, ht, -ht},

		{ht, ht, -ht},
		{ht, ht, ht},
		{ht, -ht, ht},

		// left
		{-ht, ht, -ht},
		{-ht, -ht, -ht},
		{-ht, -ht, ht},

		{-ht, -ht, ht},
		{-ht, ht, ht},
		{-ht, ht, -ht},

		// top
		{ht, ht, ht},
		{ht, ht, -ht},
		{-ht, ht, ht},

		{-ht, ht, ht},
		{ht, ht, -ht},
		{-ht, ht, -ht},

		// bottom
		{-ht, -ht, ht},
		{ht, -ht, -ht},
		{ht, -ht, ht},

		{-ht, -ht, -ht},
		{ht, -ht, -ht},
		{-ht, -ht, ht},
	}

	cubeCache[cacheKey] = points
	return points
}

func linePoints(thickness float64, length float64) []mgl64.Vec3 {
	cacheKey := genCacheKey(thickness, length)
	if _, ok := lineCache[cacheKey]; ok {
		return lineCache[cacheKey]
	}

	var ht float64 = thickness / 2
	linePoints := []mgl64.Vec3{
		// front
		{-ht, -ht, 0},
		{ht, -ht, 0},
		{ht, ht, 0},

		{ht, ht, 0},
		{-ht, ht, 0},
		{-ht, -ht, 0},

		// back
		{ht, ht, -length},
		{ht, -ht, -length},
		{-ht, -ht, -length},

		{-ht, -ht, -length},
		{-ht, ht, -length},
		{ht, ht, -length},

		// right
		{ht, -ht, 0},
		{ht, -ht, -length},
		{ht, ht, -length},

		{ht, ht, -length},
		{ht, ht, 0},
		{ht, -ht, 0},

		// left
		{-ht, ht, -length},
		{-ht, -ht, -length},
		{-ht, -ht, 0},

		{-ht, -ht, 0},
		{-ht, ht, 0},
		{-ht, ht, -length},

		// top
		{ht, ht, 0},
		{ht, ht, -length},
		{-ht, ht, 0},

		{-ht, ht, 0},
		{ht, ht, -length},
		{-ht, ht, -length},

		// bottom
		{-ht, -ht, 0},
		{ht, -ht, -length},
		{ht, -ht, 0},

		{-ht, -ht, -length},
		{ht, -ht, -length},
		{-ht, -ht, 0},
	}

	lineCache[cacheKey] = linePoints
	return linePoints
}

func (r *RenderSystem) drawBillboardTexture(
	texture uint32,
	length float32,
) {
	if singleSidedQuadVAO == 0 {
		vertices := []float32{
			-1 * length, -1 * length, 0, 0.0, 0.0,
			1 * length, -1 * length, 0, 1.0, 0.0,
			1 * length, 1 * length, 0, 1.0, 1.0,
			1 * length, 1 * length, 0, 1.0, 1.0,
			-1 * length, 1 * length, 0, 0.0, 1.0,
			-1 * length, -1 * length, 0, 0.0, 0.0,
		}

		var vbo, vao uint32
		apputils.GenBuffers(1, &vbo)
		gl.GenVertexArrays(1, &vao)

		gl.BindVertexArray(vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
		gl.EnableVertexAttribArray(1)

		singleSidedQuadVAO = vao
	}

	gl.BindVertexArray(singleSidedQuadVAO)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	r.iztDrawArrays(0, 6)
}

// drawHUDTextureToQuad does a shitty perspective based rendering of a flat texture
func (r *RenderSystem) drawHUDTextureToQuad(viewerContext context.ViewerContext, shader *shaders.ShaderProgram, texture uint32, hudScale float32) {
	// texture coords top left = 0,0 | bottom right = 1,1
	var vertices []float32 = []float32{
		// front
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
		1 * hudScale, -1 * hudScale, 0, 1.0, 0.0,
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
		1 * hudScale, 1 * hudScale, 0, 1.0, 1.0,
		-1 * hudScale, 1 * hudScale, 0, 0.0, 1.0,
		-1 * hudScale, -1 * hudScale, 0, 0.0, 0.0,
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	shader.Use()
	shader.SetUniformMat4("model", mgl32.Translate3D(0, 0, -2))
	shader.SetUniformMat4("view", mgl32.Ident4())
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	r.iztDrawArrays(0, 6)
}

func (r *RenderSystem) createCircleTexture(width, height int) (uint32, uint32) {
	circleTextureFn := textureFn(width, height, []int32{rendersettings.InternalTextureColorFormatRGBA}, []uint32{rendersettings.RenderFormatRGBA}, []uint32{gl.UNSIGNED_BYTE})
	fbo, textures := r.initFrameBuffer(circleTextureFn)
	return fbo, textures[0]
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

func (r *RenderSystem) initFrameBuffer(tf TextureFn) (uint32, []uint32) {
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

func (r *RenderSystem) initFrameBufferNoDepth(tf TextureFn) (uint32, []uint32) {
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

func (r *RenderSystem) createDepthTexture(width, height int) uint32 {
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

func (r *RenderSystem) drawSkybox(renderContext context.RenderContext, viewerContext context.ViewerContext) {
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

	shader := r.shaderManager.GetShaderProgram("skybox")
	shader.Use()
	var fog int32 = 0
	if r.app.RuntimeConfig().FogDensity != 0 {
		fog = 1
	}
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrixWithoutTranslation))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformInt("fog", fog)
	shader.SetUniformInt("fogDensity", r.app.RuntimeConfig().FogDensity)
	shader.SetUniformFloat("far", r.app.RuntimeConfig().Far)
	shader.SetUniformVec3("skyboxTopColor", r.app.RuntimeConfig().SkyboxTopColor)
	shader.SetUniformVec3("skyboxBottomColor", r.app.RuntimeConfig().SkyboxBottomColor)
	shader.SetUniformFloat("skyboxMixValue", r.app.RuntimeConfig().SkyboxMixValue)
	r.iztDrawArrays(0, 36)
	gl.DepthFunc(gl.LESS)
}

func (r *RenderSystem) CameraViewerContext() context.ViewerContext {
	return r.cameraViewerContext
}

// NOTE: this method should only be called from within the render loop. if the frame
// buffer is swapped, then the data in the buffer can be undefined. so, we should make
// sure this is called in the render loop and before we swap frame buffers. that said,
// this might be handled automatically by the graphics driver so it may not actually
// be necessary.
//
// some changes that i've made to attempt crash fixes is moving the picking buffer into
// a package variable outside of the getEntityByPixelPosition method. in addition, i've
// done better vao caching for our gizmos which previously were recreated whenever the
// camera moves
func (r *RenderSystem) getEntityByPixelPosition(pixelPosition mgl64.Vec2) *int {
	if r.app.Minimized() || !r.app.WindowFocused() {
		return nil
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.mainRenderFBO)
	gl.ReadBuffer(r.colorPickingAttachment)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, r.mainRenderFBO)

	_, windowHeight := r.app.WindowSize()
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	if len(pickingBuffer) == 0 {
		pickingBuffer = make([]byte, 4)
	}

	var footerSize int32 = 0
	if r.app.RuntimeConfig().UIEnabled {
		footerSize = int32(apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled))
	}

	// in OpenGL, the mouse origin is the bottom left corner, so we need to offset by the footer size if it's present
	// SDL, on the other hand, has the mouse origin in the top left corner
	var weirdOffset float32 = -1 // Weirdge
	gl.ReadPixels(int32(pixelPosition[0]), int32(windowHeight)-int32(pixelPosition[1])-footerSize+int32(weirdOffset), 1, 1, gl.RGB_INTEGER, gl.UNSIGNED_INT, gl.Ptr(pickingBuffer))

	uintID := binary.LittleEndian.Uint32(pickingBuffer)
	if uintID == settings.EmptyColorPickingID {
		return nil
	}

	id := int(uintID)
	return &id
}

func (r *RenderSystem) drawSpatialPartition(viewerContext context.ViewerContext, color mgl64.Vec3, spatialPartition *spatialpartition.SpatialPartition, thickness float64) {
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

	shader := r.shaderManager.GetShaderProgram("flat")
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	r.drawLineGroup("spatial_partition", shader, allLines, thickness, color)
}

func (r *RenderSystem) drawAABB(viewerContext context.ViewerContext, color mgl64.Vec3, aabb collider.BoundingBox, thickness float64) {
	var allLines [][2]mgl64.Vec3

	d := aabb.MaxVertex.Sub(aabb.MinVertex)
	xd := d.X()
	yd := d.Y()
	zd := d.Z()

	baseHorizontalLines := [][2]mgl64.Vec3{}
	baseHorizontalLines = append(baseHorizontalLines,
		[2]mgl64.Vec3{aabb.MinVertex, aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0})},
		[2]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0}), aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd})},
		[2]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd}), aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd})},
		[2]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd}), aabb.MinVertex},
	)

	for i := 0; i < 2; i++ {
		for _, b := range baseHorizontalLines {
			allLines = append(allLines,
				[2]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, float64(i) * yd, 0}), b[1].Add(mgl64.Vec3{0, float64(i) * yd, 0})},
			)
		}
	}

	baseVerticalLines := [][]mgl64.Vec3{}
	for i := 0; i < 2; i++ {
		baseVerticalLines = append(baseVerticalLines,
			[]mgl64.Vec3{aabb.MinVertex, aabb.MinVertex.Add(mgl64.Vec3{0, yd, 0})},
			[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0}), aabb.MinVertex.Add(mgl64.Vec3{xd, yd, 0})},
		)
	}

	for i := 0; i < 2; i++ {
		for _, b := range baseVerticalLines {
			allLines = append(allLines,
				[2]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, 0, float64(i) * zd}), b[1].Add(mgl64.Vec3{0, 0, float64(i) * zd})},
			)
		}
	}

	shader := r.shaderManager.GetShaderProgram("flat")
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	r.drawLineGroup(fmt.Sprintf("aabb_%v_%v", aabb.MinVertex, aabb.MaxVertex), shader, allLines, thickness, color)
}

func (r *RenderSystem) getCubeVAO(length float32, includeNormals bool) uint32 {
	hash := fmt.Sprintf("%.2f_%t", length, includeNormals)
	if _, ok := r.cubeVAOs[hash]; !ok {
		vao := r.initCubeVAO(length, includeNormals)
		r.cubeVAOs[hash] = vao
	}
	return r.cubeVAOs[hash]
}

func (r *RenderSystem) initCubeVAO(length float32, includeNormals bool) uint32 {
	ht := length / 2

	allVertexAttribs := []float32{
		// front
		-ht, -ht, ht, 0, 0, -1,
		ht, -ht, ht, 0, 0, -1,
		ht, ht, ht, 0, 0, -1,

		ht, ht, ht, 0, 0, -1,
		-ht, ht, ht, 0, 0, -1,
		-ht, -ht, ht, 0, 0, -1,

		// back
		ht, ht, -ht, 0, 0, 1,
		ht, -ht, -ht, 0, 0, 1,
		-ht, -ht, -ht, 0, 0, 1,

		-ht, -ht, -ht, 0, 0, 1,
		-ht, ht, -ht, 0, 0, 1,
		ht, ht, -ht, 0, 0, 1,

		// right
		ht, -ht, ht, 1, 0, 0,
		ht, -ht, -ht, 1, 0, 0,
		ht, ht, -ht, 1, 0, 0,

		ht, ht, -ht, 1, 0, 0,
		ht, ht, ht, 1, 0, 0,
		ht, -ht, ht, 1, 0, 0,

		// left
		-ht, ht, -ht, -1, 0, 0,
		-ht, -ht, -ht, -1, 0, 0,
		-ht, -ht, ht, -1, 0, 0,

		-ht, -ht, ht, -1, 0, 0,
		-ht, ht, ht, -1, 0, 0,
		-ht, ht, -ht, -1, 0, 0,

		// top
		ht, ht, ht, 0, 1, 0,
		ht, ht, -ht, 0, 1, 0,
		-ht, ht, ht, 0, 1, 0,

		-ht, ht, ht, 0, 1, 0,
		ht, ht, -ht, 0, 1, 0,
		-ht, ht, -ht, 0, 1, 0,

		// bottom
		-ht, -ht, ht, 0, -1, 0,
		ht, -ht, -ht, 0, -1, 0,
		ht, -ht, ht, 0, -1, 0,

		-ht, -ht, -ht, 0, -1, 0,
		ht, -ht, -ht, 0, -1, 0,
		-ht, -ht, ht, 0, -1, 0,
	}

	var vertices []float32

	totalVertexAttributesSize := 6
	actualVertexAttributesSize := totalVertexAttributesSize
	if !includeNormals {
		actualVertexAttributesSize -= 3
	}

	for i := 0; i < len(allVertexAttribs); i += totalVertexAttributesSize {
		for j := range actualVertexAttributesSize {
			vertices = append(vertices, allVertexAttribs[i+j])
		}
	}

	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	ptrOffset := 0
	floatSize := 4

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(actualVertexAttributesSize*floatSize), nil)
	gl.EnableVertexAttribArray(0)

	if includeNormals {
		ptrOffset += 3
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(actualVertexAttributesSize*floatSize), gl.PtrOffset(ptrOffset*floatSize))
		gl.EnableVertexAttribArray(1)
	}

	return vao
}

func calculateFrustumPoints(position mgl64.Vec3, rotation mgl64.Quat, near, far, fovX, fovY, aspectRatio float64, nearPlaneOffset float64) []mgl64.Vec3 {
	viewerViewMatrix := rotation.Mat4()

	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())
	viewMatrix := viewTranslationMatrix.Mul4(viewerViewMatrix)

	halfY := math.Tan(mgl64.DegToRad(fovY / 2))
	halfX := math.Tan(mgl64.DegToRad(fovX / 2))

	var verts []mgl64.Vec3

	corners := []float64{-1, 1}
	nearFar := []float64{near, far}
	offsets := []float64{nearPlaneOffset, 0}

	for k, distance := range nearFar {
		for _, i := range corners {
			for _, j := range corners {
				vert := viewMatrix.Mul4x1(mgl64.Vec3{i * halfX * distance, j * halfY * distance, -distance + offsets[k]}.Vec4(1)).Vec3()
				verts = append(verts, vert)
			}
		}
	}

	return verts
}

func (r *RenderSystem) iztDrawArrays(first, count int32) {
	r.app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	r.app.RuntimeConfig().DrawCount += 1
	gl.DrawArrays(gl.TRIANGLES, first, count)
}

func (r *RenderSystem) iztDrawLines(count int32) {
	r.app.RuntimeConfig().DrawCount += 1
	gl.DrawArrays(gl.LINES, 0, count)
}

func (r *RenderSystem) iztDrawElements(count int32) {
	r.app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	r.app.RuntimeConfig().DrawCount += 1
	gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, nil)
}

// setup reusale circle textures
func (r *RenderSystem) initializeCircleTextures() {
	gl.Viewport(0, 0, 1024, 1024)
	shaderManager := r.shaderManager
	shader := shaderManager.GetShaderProgram("unit_circle")
	shader.Use()

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.redCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{1, 0, 0, 1})
	r.drawCircle()

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.greenCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{0, 1, 0, 1})
	r.drawCircle()

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.blueCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{0, 0, 1, 1})
	r.drawCircle()

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.yellowCircleFB)
	gl.ClearColor(0, 0.5, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	shader.SetUniformVec4("color", mgl32.Vec4{1, 1, 0, 1})
	r.drawCircle()
}

func CalculateMenuBarHeight() float32 {
	style := imgui.CurrentStyle()
	return settings.FontSize + style.FramePadding().Y*2
}

func (r *RenderSystem) GameWindowSize() (int, int) {
	menuBarSize := CalculateMenuBarHeight()
	footerSize := apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled)

	windowWidth, windowHeight := r.app.WindowSize()

	width := windowWidth
	height := windowHeight - int(menuBarSize) - int(footerSize)

	if r.app.RuntimeConfig().UIEnabled {
		width = int(math.Ceil(float64(1-uiWidthRatio) * float64(windowWidth)))
	}

	return width, height
}
