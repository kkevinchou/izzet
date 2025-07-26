package render

import (
	"fmt"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/entities"
)

const maxHemisphereSamples int = 64
const maxSSAONoise int = 16

func (r *RenderSystem) drawSSAO(viewerContext ViewerContext, lightContext LightContext, renderContext RenderContext, renderableEntities []*entities.Entity) {
	gl.Viewport(0, 0, int32(renderContext.Width()), int32(renderContext.Height()))
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	shader := r.shaderManager.GetShaderProgram("ssao")
	shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, r.gPositionTexture)
	shader.SetUniformInt("gPosition", 0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, r.gNormalTexture)
	shader.SetUniformInt("gNormal", 1)

	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, r.ssaoNoiseTexture)
	shader.SetUniformInt("texNoise", 2)

	for i := range maxHemisphereSamples {
		shader.SetUniformVec3(fmt.Sprintf("samples[%d]", i), r.ssaoSamples[i])
	}

	shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	shader.SetUniformFloat("radius", r.app.RuntimeConfig().SSAORadius)
	shader.SetUniformFloat("bias", r.app.RuntimeConfig().SSAOBias)

	gl.BindVertexArray(r.ndcQuadVAO)
	r.iztDrawArrays(0, 6)
}

func (r *RenderSystem) renderSSAOTextures() {
	gl.Viewport(0, 0, 1024, 1024)

	var noiseTexture uint32
	gl.GenTextures(1, &noiseTexture)
	gl.BindTexture(gl.TEXTURE_2D, noiseTexture)

	noiseFloats := ssaoNoise()
	gl.TexImage2D(gl.TEXTURE_2D, 0, internalTextureColorFormat16RGBA, 4, 4, 0, renderFormatRGB, gl.FLOAT, gl.Ptr(noiseFloats))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	r.ssaoNoiseTexture = noiseTexture

	r.ssaoSamples = randomHemisphereVectors()
}

func randomHemisphereVectors() [maxHemisphereSamples]mgl32.Vec3 {
	var result [maxHemisphereSamples]mgl32.Vec3

	for i := range maxHemisphereSamples {
		v := mgl32.Vec3{
			rand.Float32()*2 - 1,
			rand.Float32()*2 - 1,
			rand.Float32(),
		}
		scale := float32(i) / 64
		scale = lerp(0.1, 1, scale*scale)
		v = v.Normalize()
		v = v.Mul(rand.Float32() * scale)
		result[i] = v
	}

	return result
}

func lerp(a, b, f float32) float32 {
	return a + f*(b-a)
}

func ssaoNoise() []mgl32.Vec3 {
	var result []mgl32.Vec3
	for _ = range maxSSAONoise {
		v := mgl32.Vec3{
			rand.Float32()*2 - 1,
			rand.Float32()*2 - 1,
			0,
		}
		result = append(result, v)

	}
	return result
}
