package render

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/app/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/navmesh"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/assets"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/spatialpartition"
	"github.com/kkevinchou/kitolib/utils"
)

var lineCache map[string][]mgl64.Vec3

func init() {
	lineCache = map[string][]mgl64.Vec3{}
}

func genLineKey(thickness, length float64) string {
	return fmt.Sprintf("%.3f_%.3f", thickness, length)
}

// drawTris draws a list of triangles in winding order. each triangle is defined with 3 consecutive points
func (r *Renderer) drawTris(points []mgl64.Vec3) {
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

	gl.BindVertexArray(vao)
	r.iztDrawArrays(0, int32(len(vertices)))
}

var navMeshTrisVAO uint32
var navMeshVBO uint32
var lastVoxelCount = 0
var lastVertexCount = 0
var ResetNavMeshVAO bool = false

var lastMeshUpdate time.Time = time.Now()

// drawTris draws a list of triangles in winding order. each triangle is defined with 3 consecutive points
func (r *Renderer) drawNavMeshTris(viewerContext ViewerContext, navmesh *navmesh.NavigationMesh) {
	if navmesh.VoxelCount() != lastVoxelCount || ResetNavMeshVAO {
		if time.Since(lastMeshUpdate) > 5*time.Second || ResetNavMeshVAO {
			ResetNavMeshVAO = false
			vaos := []uint32{navMeshTrisVAO}
			gl.DeleteVertexArrays(1, &vaos[0])
			vbos := []uint32{navMeshVBO}
			gl.DeleteBuffers(1, &vbos[0])

			vertices := r.generateNavMeshVertexAttributes(navmesh)

			var vbo, vao uint32
			apputils.GenBuffers(1, &vbo)
			gl.GenVertexArrays(1, &vao)

			gl.BindVertexArray(vao)
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

			var stride int32 = 9

			gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride*4, nil)
			gl.EnableVertexAttribArray(0)

			gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride*4, gl.PtrOffset(3*4))
			gl.EnableVertexAttribArray(1)

			gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride*4, gl.PtrOffset(6*4))
			gl.EnableVertexAttribArray(2)

			navMeshTrisVAO = vao
			navMeshVBO = vbo
			lastVoxelCount = navmesh.VoxelCount()
			lastVertexCount = len(vertices) / int(stride)
			lastMeshUpdate = time.Now()
		}
	}

	gl.BindVertexArray(navMeshTrisVAO)
	r.iztDrawArrays(0, int32(lastVertexCount))
}

func (r *Renderer) generateNavMeshVertexAttributes(navmesh *navmesh.NavigationMesh) []float32 {
	delta := navmesh.Volume.MaxVertex.Sub(navmesh.Volume.MinVertex)
	voxelDimension := navmesh.VoxelDimension()
	var runs [3]int = [3]int{int(delta[0] / voxelDimension), int(delta[1] / voxelDimension), int(delta[2] / voxelDimension)}

	voxelField := navmesh.VoxelField()
	var vertexAttributes []float32

	for i := 0; i < runs[0]; i++ {
		for j := 0; j < runs[1]; j++ {
			for k := 0; k < runs[2]; k++ {
				voxel := voxelField[i][j][k]
				if !voxel.Filled {
					continue
				}

				bb := collider.BoundingBox{
					MinVertex: navmesh.Volume.MinVertex.Add(mgl64.Vec3{float64(i), float64(j), float64(k)}.Mul(voxelDimension)),
					MaxVertex: navmesh.Volume.MinVertex.Add(mgl64.Vec3{float64(i + 1), float64(j + 1), float64(k + 1)}.Mul(voxelDimension)),
				}
				vertexAttributes = append(vertexAttributes, r.generateVoxelVertexAttributes(voxel, voxelField, bb)...)
			}
		}
	}
	return vertexAttributes
}

func (r *Renderer) generateVoxelVertexAttributes(voxel navmesh.Voxel, voxelField [][][]navmesh.Voxel, bb collider.BoundingBox) []float32 {
	min := bb.MinVertex
	max := bb.MaxVertex
	delta := max.Sub(min)
	var vertexAttributes []float32

	verts := []mgl64.Vec3{
		// top
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		max,
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),

		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),
		max,

		// bottom
		min,
		min.Add(mgl64.Vec3{delta[0], 0, 0}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),

		min,
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),
		min.Add(mgl64.Vec3{0, 0, delta[2]}),

		// left
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min,
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		min,
		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		// right
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, 0}),

		min.Add(mgl64.Vec3{delta[0], 0, 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),

		// front
		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),

		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		// back
		min,
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], 0, 0}),

		min,
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
	}

	normals := []mgl64.Vec3{
		// top
		mgl64.Vec3{0, 1, 0},
		// bottom
		mgl64.Vec3{0, -1, 0},
		// left
		mgl64.Vec3{-1, 0, 0},
		// right
		mgl64.Vec3{1, 0, 0},
		// front
		mgl64.Vec3{0, 0, 1},
		// back
		mgl64.Vec3{0, 0, -1},
	}

	// color := []float32{3.0 / 255, 185.0 / 255, 5.0 / 255}
	// if voxel.DistanceField == 0 {
	// 	color = []float32{0.8, 0, 0}
	// }

	color := mgl32.Vec3{1.0, 0, 0}

	// if len(voxel.Neighbors) == 4 {
	// 	color = mgl32.Vec3{0, 1, 0}
	// } else if voxel.DistanceField < navmesh.MaxDistanceFieldValue {
	if voxel.DistanceField < navmesh.MaxDistanceFieldValue {
		hsv := mgl32.Vec3{0, 0, float32(voxel.DistanceField) / 100}
		color = HSVtoRGB(hsv)

		if voxel.X == int(r.app.RuntimeConfig().VoxelHighlightX) && voxel.Z == int(r.app.RuntimeConfig().VoxelHighlightZ) && voxel.Y < 50 {
			r.app.RuntimeConfig().VoxelHighlightDistanceField = float32(voxel.DistanceField)
			r.app.RuntimeConfig().VoxelHighlightRegionID = voxel.RegionID
			color = mgl32.Vec3{10, 10, 10}
		} else if voxel.Seed {
			color = mgl32.Vec3{1, 0, 1}
		} else if r.app.RuntimeConfig().NavMeshHSV {
			if voxel.RegionID != -1 && voxel.RegionID <= int(r.app.RuntimeConfig().NavMeshRegionIDThreshold) && voxel.DistanceField >= float64(r.app.RuntimeConfig().NavMeshDistanceFieldThreshold) {
				// if voxel.RegionID != -1 {
				hsv = mgl32.Vec3{float32((voxel.RegionID * int(r.app.RuntimeConfig().HSVOffset)) % 255), 1, 1}
				color = HSVtoRGB(hsv)
			}
		}

		if voxel.DEBUGCOLORFACTOR != nil {
			color = color.Mul(*voxel.DEBUGCOLORFACTOR)
		}
	}

	// if voxel.DistanceField == 0 {
	// 	colorVal := float32(1)
	// 	color = []float32{colorVal, colorVal, colorVal}
	// }

	for i := 0; i < len(verts); i++ {
		vertexAttributes = append(vertexAttributes,
			float32(verts[i].X()),
			float32(verts[i].Y()),
			float32(verts[i].Z()),
			float32(normals[i/6].X()),
			float32(normals[i/6].Y()),
			float32(normals[i/6].Z()),
			color[0],
			color[1],
			color[2],
		)
	}

	return vertexAttributes
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

type RenderData struct {
	Primitive   *modelspec.PrimitiveSpecification
	Transform   mgl32.Mat4
	VAO         uint32
	GeometryVAO uint32
}

// func getRenderData(modelLibrary *modellibrary.ModelLibrary, entity *entities.Entity) []RenderData {
// 	var result []RenderData

// 	if entity.MeshComponent != nil {
// 		primitives := modelLibrary.GetPrimitives(entity.MeshComponent.MeshHandle)
// 		for _, p := range primitives {
// 			result = append(result, RenderData{
// 				Primitive:   p.Primitive,
// 				Transform:   utils.Mat4F64ToF32(entity.MeshComponent.Transform),
// 				VAO:         p.VAO,
// 				GeometryVAO: p.GeometryVAO,
// 			})
// 		}
// 	}

// 	return result
// }

func (r *Renderer) drawModel(
	viewerContext ViewerContext,
	lightContext LightContext,
	shadowMap *ShadowMap,
	shader *shaders.ShaderProgram,
	assetManager *assets.AssetManager,
	animationPlayer *animation.AnimationPlayer,
	modelMatrix mgl64.Mat4,
	pointLightDepthCubeMap uint32,
	entityID int,
	material *entities.MaterialComponent,
	modelLibrary *modellibrary.ModelLibrary,
	entity *entities.Entity,
) {

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
	primitives := modelLibrary.GetPrimitives(entity.MeshComponent.MeshHandle)
	for _, p := range primitives {
		if material == nil && p.Primitive.PBRMaterial == nil {
			shader.SetUniformInt("hasPBRBaseColorTexture", 0)
			shader.SetUniformVec3("albedo", mgl32.Vec3{255.0 / 255, 28.0 / 255, 217.0 / 121.0})
			shader.SetUniformFloat("roughness", 0.85)
			shader.SetUniformFloat("metallic", 0.1)

			gl.ActiveTexture(gl.TEXTURE0)
			var textureID uint32
			textureName := settings.DefaultTexture
			texture := assetManager.GetTexture(textureName)
			textureID = texture.ID
			gl.BindTexture(gl.TEXTURE_2D, textureID)
		} else if material == nil {
			primitiveMaterial := p.Primitive.PBRMaterial.PBRMetallicRoughness
			shader.SetUniformInt("colorTextureCoordIndex", int32(primitiveMaterial.BaseColorTextureCoordsIndex))

			if primitiveMaterial.BaseColorTextureIndex != nil {
				shader.SetUniformInt("hasPBRBaseColorTexture", 1)
			} else {
				shader.SetUniformInt("hasPBRBaseColorTexture", 0)
			}

			shader.SetUniformVec3("albedo", primitiveMaterial.BaseColorFactor.Vec3())
			if r.app.RuntimeConfig().MaterialOverride {
				shader.SetUniformFloat("roughness", r.app.RuntimeConfig().Roughness)
				shader.SetUniformFloat("metallic", r.app.RuntimeConfig().Metallic)
			} else {
				shader.SetUniformFloat("roughness", primitiveMaterial.RoughnessFactor)
				shader.SetUniformFloat("metallic", primitiveMaterial.MetalicFactor)
			}

			// main diffuse texture
			gl.ActiveTexture(gl.TEXTURE0)
			var textureID uint32
			textureName := settings.DefaultTexture
			if p.Primitive.TextureName() != "" {
				textureName = p.Primitive.TextureName()
			}
			texture := assetManager.GetTexture(textureName)
			textureID = texture.ID
			gl.BindTexture(gl.TEXTURE_2D, textureID)
		} else {
			if material.Invisible {
				return
			}
			var color mgl32.Vec3 = material.PBR.Diffuse
			shader.SetUniformInt("hasPBRBaseColorTexture", 0)
			shader.SetUniformVec3("albedo", color.Mul(material.PBR.DiffuseIntensity))
			shader.SetUniformFloat("roughness", material.PBR.Roughness)
			shader.SetUniformFloat("metallic", material.PBR.Metallic)
		}
		shader.SetUniformFloat("ao", 1.0)

		modelMat := utils.Mat4F64ToF32(modelMatrix).Mul4(utils.Mat4F64ToF32(entity.MeshComponent.Transform))
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

func toRadians(degrees float64) float64 {
	return degrees / 180 * math.Pi
}

func (r *Renderer) drawLines(viewerContext ViewerContext, shader *shaders.ShaderProgram, lines [][]mgl64.Vec3, thickness float64, color mgl64.Vec3) {
	var points []mgl64.Vec3
	for _, line := range lines {
		start := line[0]
		end := line[1]
		length := end.Sub(start).Len()

		dir := end.Sub(start).Normalize()
		q := mgl64.QuatBetweenVectors(mgl64.Vec3{0, 0, -1}, dir)

		for _, dp := range linePoints(thickness, length) {
			newEnd := q.Rotate(dp).Add(start)
			points = append(points, newEnd)
		}
	}
	shader.SetUniformVec3("color", utils.Vec3F64ToF32(color))
	shader.SetUniformFloat("intensity", 1.0)
	r.drawTris(points)
}

func cubeLines(length float64) [][]mgl64.Vec3 {
	directions := [][]float64{
		[]float64{-1, 1, 0.5},
		[]float64{-1, -1, 0.5},
		[]float64{1, -1, 0.5},
		[]float64{1, 1, 0.5},
	}

	position := mgl64.Vec3{}
	var lines [][]mgl64.Vec3
	var frontPoints []mgl64.Vec3

	// front points
	for _, direction := range directions {
		point := position.Add(mgl64.Vec3{direction[0] * length / 2, direction[1] * length / 2, direction[2] * length})
		frontPoints = append(frontPoints, point)
	}
	for i := range frontPoints {
		line := []mgl64.Vec3{frontPoints[i], frontPoints[(i+1)%len(frontPoints)]}
		lines = append(lines, line)
	}

	// back points
	var backPoints []mgl64.Vec3
	for _, point := range frontPoints {
		backPoints = append(backPoints, point.Add(mgl64.Vec3{0, 0, -length}))
	}
	for i := range backPoints {
		line := []mgl64.Vec3{backPoints[i], backPoints[(i+1)%len(backPoints)]}
		lines = append(lines, line)
	}

	// connect front and back
	for i := range frontPoints {
		line := []mgl64.Vec3{frontPoints[i], backPoints[i]}
		lines = append(lines, line)
	}

	return lines
}

func linePoints(thickness float64, length float64) []mgl64.Vec3 {
	cacheKey := genLineKey(thickness, length)
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

func (r *Renderer) drawWithNDC(shaderManager *shaders.ShaderManager) {
	// triangle
	// var vertices []float32 = []float32{
	// 	-0.5, -0.5, 1,
	// 	0.5, -0.5, 1,
	// 	0.0, 0.5, 1,
	// }

	var back float32 = 1

	// full screen
	var vertices []float32 = []float32{
		-1, 1, back,
		-1, -1, back,
		1, -1, back,

		1, -1, back,
		1, 1, back,
		-1, 1, back,
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

	shader := shaderManager.GetShaderProgram("skybox")
	shader.Use()
	var fog int32 = 0
	if r.app.RuntimeConfig().FogDensity != 0 {
		fog = 1
	}
	shader.SetUniformInt("fog", fog)
	shader.SetUniformInt("fogDensity", r.app.RuntimeConfig().FogDensity)
	shader.SetUniformFloat("far", r.app.RuntimeConfig().Far)
	r.iztDrawArrays(0, int32(len(vertices)))
}

var singleSidedQuadVAO uint32

func (r *Renderer) drawBillboardTexture(
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
func (r *Renderer) drawHUDTextureToQuad(viewerContext ViewerContext, shader *shaders.ShaderProgram, texture uint32, hudScale float32) {
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

func (r *Renderer) initFrameBufferSingleColorAttachment(width, height int, internalFormat int32, format uint32) (uint32, uint32) {
	fbo, textures := r.initFrameBuffer(width, height, []int32{internalFormat}, []uint32{format})
	return fbo, textures[0]
}

func (r *Renderer) initFrameBuffer(width int, height int, internalFormat []int32, format []uint32) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var textures []uint32
	var drawBuffers []uint32

	colorBufferCount := len(internalFormat)

	for i := 0; i < colorBufferCount; i++ {
		var texture uint32
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)

		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat[i],
			int32(width), int32(height), 0, format[i], gl.UNSIGNED_BYTE, nil)

		gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachment, gl.TEXTURE_2D, texture, 0)

		textures = append(textures, texture)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(colorBufferCount), &drawBuffers[0])

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

func (r *Renderer) initFBOAndTexture(width, height int) (uint32, uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, internalTextureColorFormat,
		int32(width), int32(height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	drawBuffers := []uint32{gl.COLOR_ATTACHMENT0}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)
	gl.DrawBuffers(1, &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, texture
}

func (r *Renderer) clearMainFrameBuffer(renderContext RenderContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.renderFBO)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) drawSkybox(renderContext RenderContext) {
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	r.drawWithNDC(r.shaderManager)
}

func (r *Renderer) CameraViewerContext() ViewerContext {
	return r.cameraViewerContext
}

func (r *Renderer) GetEntityByPixelPosition(pixelPosition mgl64.Vec2) *int {
	if r.app.Minimized() || !r.app.WindowFocused() {
		return nil
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.renderFBO)
	gl.ReadBuffer(r.colorPickingAttachment)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, r.renderFBO)

	_, windowHeight := r.app.WindowSize()
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	data := make([]byte, 4)

	var footerSize int32 = 0
	if r.app.RuntimeConfig().UIEnabled {
		footerSize = int32(apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled))
	}

	// in OpenGL, the mouse origin is the bottom left corner, so we need to offset by the footer size if it's present
	// SDL, on the other hand, has the mouse origin in the top left corner
	var weirdOffset float32 = -1 // Weirdge
	gl.ReadPixels(int32(pixelPosition[0]), int32(windowHeight)-int32(pixelPosition[1])-footerSize+int32(weirdOffset), 1, 1, gl.RGB_INTEGER, gl.UNSIGNED_INT, gl.Ptr(data))

	uintID := binary.LittleEndian.Uint32(data)
	if uintID == settings.EmptyColorPickingID {
		return nil
	}

	id := int(uintID)
	return &id
}

// func (r *Renderer) ReadAllPixels() {
// 	// start := time.Now()

// 	// Specify the texture target
// 	gl.BindTexture(gl.TEXTURE_2D, r.cameraDepthTexture)

// 	// Allocate memory to store the pixel data
// 	// var internalFormat uint32 = gl.RGBA    // The format of the texture data (e.g., gl.RGBA)
// 	// var dataType uint32 = gl.UNSIGNED_BYTE // The data type of the texture (e.g., gl.UNSIGNED_BYTE)

// 	var internalFormat uint32 = gl.DEPTH_COMPONENT // The format of the texture data (e.g., gl.RGBA)
// 	var dataType uint32 = gl.FLOAT                 // The data type of the texture (e.g., gl.UNSIGNED_BYTE)

// 	// Calculate the size of the buffer
// 	bufferSize := int(r.windowWidth * r.windowHeight * 4) // Assuming 4 components per pixel (RGBA)

// 	// Allocate memory for the pixel data
// 	pixelData := make([]byte, bufferSize)

// 	// Read the texture data into the pixelData slice
// 	gl.GetTexImage(gl.TEXTURE_2D, 0, internalFormat, dataType, gl.Ptr(pixelData))

// 	// Now, pixelData contains the pixel data of the texture.
// 	// You can process it or save it as needed.

// 	start := time.Now()
// 	// Print a few pixels as an example
// 	for i := 0; i < r.windowWidth*r.windowHeight*4-4; i++ {
// 		r := pixelData[i]
// 		g := pixelData[i+1]
// 		b := pixelData[i+2]
// 		a := pixelData[i+3]

// 		if r != 0 {
// 			a := 5
// 			_ = a
// 		}
// 		if g != 0 {
// 			a := 5
// 			_ = a
// 		}
// 		if b != 0 {
// 			a := 5
// 			_ = a
// 		}
// 		if a != 0 {
// 			a := 5
// 			_ = a
// 		}
// 		// fmt.Printf("Pixel %d: R=%d, G=%d, B=%d, A=%d\n", i/4, r, g, b, a)
// 	}
// 	fmt.Println(time.Since(start))

// 	// fmt.Println(time.Since(start))
// }

var (
	spatialPartitionLineCache [][]mgl64.Vec3
)

func (r *Renderer) drawSpatialPartition(viewerContext ViewerContext, color mgl64.Vec3, spatialPartition *spatialpartition.SpatialPartition, thickness float64) {
	var allLines [][]mgl64.Vec3

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
					[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, float64(i * spatialPartition.PartitionDimension), 0}), b[1].Add(mgl64.Vec3{0, float64(i * spatialPartition.PartitionDimension), 0})},
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
					[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, 0, float64(i * spatialPartition.PartitionDimension)}), b[1].Add(mgl64.Vec3{0, 0, float64(i * spatialPartition.PartitionDimension)})},
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

	r.drawLines(
		viewerContext,
		shader,
		allLines,
		thickness,
		color,
	)
}

func (r *Renderer) drawAABB(viewerContext ViewerContext, color mgl64.Vec3, aabb collider.BoundingBox, thickness float64) {
	var allLines [][]mgl64.Vec3

	d := aabb.MaxVertex.Sub(aabb.MinVertex)
	xd := d.X()
	yd := d.Y()
	zd := d.Z()

	baseHorizontalLines := [][]mgl64.Vec3{}
	baseHorizontalLines = append(baseHorizontalLines,
		[]mgl64.Vec3{aabb.MinVertex, aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0})},
		[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, 0}), aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd})},
		[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{xd, 0, zd}), aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd})},
		[]mgl64.Vec3{aabb.MinVertex.Add(mgl64.Vec3{0, 0, zd}), aabb.MinVertex},
	)

	for i := 0; i < 2; i++ {
		for _, b := range baseHorizontalLines {
			allLines = append(allLines,
				[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, float64(i) * yd, 0}), b[1].Add(mgl64.Vec3{0, float64(i) * yd, 0})},
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
				[]mgl64.Vec3{b[0].Add(mgl64.Vec3{0, 0, float64(i) * zd}), b[1].Add(mgl64.Vec3{0, 0, float64(i) * zd})},
			)
		}
	}

	shader := r.shaderManager.GetShaderProgram("flat")
	shader.Use()
	shader.SetUniformMat4("model", utils.Mat4F64ToF32(mgl64.Ident4()))
	shader.SetUniformMat4("view", utils.Mat4F64ToF32(viewerContext.InverseViewMatrix))
	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))

	r.drawLines(
		viewerContext,
		shader,
		allLines,
		thickness,
		color,
	)
}

func (r *Renderer) getCubeVAO(length float32) uint32 {
	if vao, ok := r.cubeVAOs[length]; !ok {
		vao = r.initCubeVAO(length)
		r.cubeVAOs[length] = vao
	}
	return r.cubeVAOs[length]
}

func (r *Renderer) initCubeVAO(length float32) uint32 {
	ht := length / 2

	vertices := []float32{
		// front
		-ht, -ht, ht,
		ht, -ht, ht,
		ht, ht, ht,

		ht, ht, ht,
		-ht, ht, ht,
		-ht, -ht, ht,

		// back
		ht, ht, -ht,
		ht, -ht, -ht,
		-ht, -ht, -ht,

		-ht, -ht, -ht,
		-ht, ht, -ht,
		ht, ht, -ht,

		// right
		ht, -ht, ht,
		ht, -ht, -ht,
		ht, ht, -ht,

		ht, ht, -ht,
		ht, ht, ht,
		ht, -ht, ht,

		// left
		-ht, ht, -ht,
		-ht, -ht, -ht,
		-ht, -ht, ht,

		-ht, -ht, ht,
		-ht, ht, ht,
		-ht, ht, -ht,

		// top
		ht, ht, ht,
		ht, ht, -ht,
		-ht, ht, ht,

		-ht, ht, ht,
		ht, ht, -ht,
		-ht, ht, -ht,

		// bottom
		-ht, -ht, ht,
		ht, -ht, -ht,
		ht, -ht, ht,

		-ht, -ht, -ht,
		ht, -ht, -ht,
		-ht, -ht, ht,
	}
	var vbo, vao uint32
	apputils.GenBuffers(1, &vbo)
	gl.GenVertexArrays(1, &vao)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	// gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, nil)
	// gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(vao)
	r.iztDrawArrays(0, int32(len(vertices))/3)

	return vao
}

func (r *Renderer) initTriangleVAO(v1, v2, v3 mgl64.Vec3) uint32 {
	vertices := []float32{
		float32(v1.X()), float32(v1.Y()), float32(v1.Z()),
		float32(v2.X()), float32(v2.Y()), float32(v2.Z()),
		float32(v3.X()), float32(v3.Y()), float32(v3.Z()),

		float32(v1.X()), float32(v1.Y()), float32(v1.Z()),
		float32(v3.X()), float32(v3.Y()), float32(v3.Z()),
		float32(v2.X()), float32(v2.Y()), float32(v2.Z()),
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
	r.iztDrawArrays(0, int32(len(vertices))/3)

	return vao
}

func calculateFrustumPoints(position mgl64.Vec3, rotation mgl64.Quat, near, far, fovX, fovY, aspectRatio float64, nearPlaneOffset float64, farPlaneScaleFactor float64) []mgl64.Vec3 {
	viewerViewMatrix := rotation.Mat4()

	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())
	viewMatrix := viewTranslationMatrix.Mul4(viewerViewMatrix)

	halfY := math.Tan(mgl64.DegToRad(fovY / 2))
	halfX := math.Tan(mgl64.DegToRad(fovX / 2))

	var verts []mgl64.Vec3

	corners := []float64{-1, 1}
	nearFar := []float64{near, far * farPlaneScaleFactor}
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

func (r *Renderer) iztDrawArrays(first, count int32) {
	r.app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	r.app.RuntimeConfig().DrawCount += 1
	gl.DrawArrays(gl.TRIANGLES, first, count)
}

func (r *Renderer) iztDrawElements(count int32) {
	r.app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	r.app.RuntimeConfig().DrawCount += 1
	gl.DrawElements(gl.TRIANGLES, count, gl.UNSIGNED_INT, nil)
}

// setup reusale circle textures
func (r *Renderer) initializeCircleTextures() {
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

func CalculateMenuBarSize() float32 {
	style := imgui.CurrentStyle()
	return settings.FontSize + style.FramePadding().Y*2
}

func (r *Renderer) ConfigureUI() {
	r.ReinitializeFrameBuffers()
	r.contentBrowserHeight = apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled)
}

func (r *Renderer) GameWindowSize() (int, int) {
	return r.gameWindowWidth, r.gameWindowHeight
}
