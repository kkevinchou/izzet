package render

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type ViewerContext struct {
	Position mgl32.Vec3
	Rotation mgl32.Quat

	InverseViewMatrix mgl32.Mat4
	ProjectionMatrix  mgl32.Mat4
}

type LightContext struct {
	LightSpaceMatrix mgl32.Mat4
	Lights           []*entities.Entity
	PointLights      []*entities.Entity
}

type RenderContext struct {
	width       int
	height      int
	aspectRatio float32
	fovX        float32
	fovY        float32
}

func NewRenderContext(width, height int, fovX float32) RenderContext {
	aspectRatio := float32(width) / float32(height)
	return RenderContext{
		width:       width,
		height:      height,
		aspectRatio: aspectRatio,
		fovX:        fovX,
		fovY:        mgl32.RadToDeg(2 * float32(math.Atan(math.Tan(float64(mgl32.DegToRad(fovX)/2)/float64(aspectRatio))))),
	}
}

func (r RenderContext) Width() int {
	return r.width
}
func (r RenderContext) Height() int {
	return r.height
}
func (r RenderContext) AspectRatio() float32 {
	return r.aspectRatio
}
func (r RenderContext) FovX() float32 {
	return r.fovX
}
func (r RenderContext) FovY() float32 {
	return r.fovY
}
