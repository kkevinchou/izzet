package renderpass

import (
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/shaders"
)

// RenderPass is a single step in the frame‐render pipeline.
type RenderPass interface {
	// Name helps with logging or debugging
	Name() string

	// Init is called once at startup (or when switching pipelines)
	Init(app renderiface.App, shaders *shaders.ShaderManager) error

	// Resize is called whenever the viewport changes size
	Resize(width, height int)

	// Render executes the pass. It may read from
	// previous-output textures and write into its own FBO.
	Render(ctx context.RenderContext)
}

func Init() {
	// gl.BindFramebuffer(...)
	// gl.Viewport(...)
	// gl.Clear(...)
	// shader := r.shaderManager.GetShaderProgram("foo")
	// shader.Use()
	// r.iztDrawArrays(…)
}
