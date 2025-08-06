package renderpass

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/utils"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
	"github.com/kkevinchou/kitolib/shaders"
)

const maxHemisphereSamples int = 64
const maxSSAONoise int = 16

type SSAORenderPass struct {
	App    renderiface.App
	shader *shaders.ShaderProgram

	ssaoNoiseTexture uint32
	ssaoSamples      [maxHemisphereSamples]mgl32.Vec3
}

func NewSSAOPass(app renderiface.App, sm *shaders.ShaderManager) *SSAORenderPass {
	return &SSAORenderPass{App: app, shader: sm.GetShaderProgram("ssao")}
}

func (p *SSAORenderPass) Init(width, height int, ctx *context.RenderPassContext) error {
	ssaoTextureFn := textureFn(width, height, []int32{gl.RED}, []uint32{gl.RED}, []uint32{gl.FLOAT})
	ssaoFBO, textures := initFrameBufferNoDepth(ssaoTextureFn)
	ctx.SSAOFBO = ssaoFBO
	ctx.SSAOTexture = textures[0]
	p.ssaoSamples = randomHemisphereVectors()
	p.setupSSAOTextures()
	return nil
}

func (p *SSAORenderPass) setupSSAOTextures() {
	// no frame buffer needed since we're just directly writing to the texture and not rendering
	gl.Viewport(0, 0, 1024, 1024)

	var noiseTexture uint32
	gl.GenTextures(1, &noiseTexture)
	gl.BindTexture(gl.TEXTURE_2D, noiseTexture)

	noiseFloats := ssaoNoise()
	gl.TexImage2D(gl.TEXTURE_2D, 0, rendersettings.InternalTextureColorFormat16RGBA, 4, 4, 0, rendersettings.RenderFormatRGB, gl.FLOAT, gl.Ptr(noiseFloats))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	p.ssaoNoiseTexture = noiseTexture
}

func (p *SSAORenderPass) Resize(width, height int, ctx *context.RenderPassContext) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, ctx.SSAOFBO)

	ssaoTextureFn := textureFn(width, height, []int32{gl.RED}, []uint32{gl.RED}, []uint32{gl.FLOAT})
	_, _, textures := ssaoTextureFn()
	ctx.SSAOTexture = textures[0]
}

// TODO - in general could make some better help methods to set uniforms
// TODO - do the entity query ourselves? take in a world?
func (p *SSAORenderPass) Render(ctx context.RenderContext, rctx *context.RenderPassContext, viewerContext context.ViewerContext) {
	mr := p.App.MetricsRegistry()
	start := time.Now()

	gl.BindFramebuffer(gl.FRAMEBUFFER, rctx.SSAOFBO)
	gl.Viewport(0, 0, int32(ctx.Width()), int32(ctx.Height()))
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	p.shader.Use()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, rctx.GPositionTexture)
	p.shader.SetUniformInt("gPosition", 0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, rctx.GNormalTexture)
	p.shader.SetUniformInt("gNormal", 1)

	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, p.ssaoNoiseTexture)
	p.shader.SetUniformInt("texNoise", 2)

	for i := range maxHemisphereSamples {
		p.shader.SetUniformVec3(fmt.Sprintf("samples[%d]", i), p.ssaoSamples[i])
	}

	p.shader.SetUniformMat4("projection", utils.Mat4F64ToF32(viewerContext.ProjectionMatrix))
	p.shader.SetUniformFloat("radius", p.App.RuntimeConfig().SSAORadius)
	p.shader.SetUniformFloat("bias", p.App.RuntimeConfig().SSAOBias)

	gl.BindVertexArray(ndcQuadVAO())
	iztDrawArrays(p.App, 0, 6)

	mr.Inc("render_ssao", float64(time.Since(start).Milliseconds()))
}

func ssaoNoise() []mgl32.Vec3 {
	var result []mgl32.Vec3
	for range maxSSAONoise {
		v := mgl32.Vec3{
			rand.Float32()*2 - 1,
			rand.Float32()*2 - 1,
			0,
		}
		result = append(result, v)

	}
	return result
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

func ndcQuadVAO() uint32 {
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

func lerp(a, b, f float32) float32 {
	return a + f*(b-a)
}
