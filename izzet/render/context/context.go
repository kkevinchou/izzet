package context

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entities"
)

type ViewerContext struct {
	Position mgl64.Vec3
	Rotation mgl64.Quat

	InverseViewMatrix                   mgl64.Mat4
	InverseViewMatrixWithoutTranslation mgl64.Mat4
	ProjectionMatrix                    mgl64.Mat4
}

type LightContext struct {
	LightSpaceMatrix mgl64.Mat4
	Lights           []*entities.Entity
	PointLights      []*entities.Entity
}

type RenderContext struct {
	width       int
	height      int
	aspectRatio float64
	fovX        float64
	fovY        float64

	BatchRenders []assets.Batch
}

// intermediate rendering properties
type RenderPassContext struct {
	// entities
	ShadowCastingEntities []*entities.Entity
	RenderableEntities    []*entities.Entity

	// Gpass
	GeometryFBO      uint32
	GPositionTexture uint32
	GNormalTexture   uint32
	GColorTexture    uint32

	// SSAO
	SSAOFBO     uint32
	SSAOTexture uint32

	// SAO Blur
	SSAOBlurFBO     uint32
	SSAOBlurTexture uint32

	// Camera Depth
	CameraDepthFBO     uint32
	CameraDepthTexture uint32

	// Point Light
	PointLightFBO     uint32
	PointLightTexture uint32

	// Shadow Map
	ShadowMapFBO     uint32
	ShadowMapTexture uint32
}

func NewRenderContext(width, height int, fovX float64) RenderContext {
	aspectRatio := float64(width) / float64(height)
	return RenderContext{
		width:       width,
		height:      height,
		aspectRatio: aspectRatio,
		fovX:        fovX,
		fovY:        mgl64.RadToDeg(2 * math.Atan(math.Tan(mgl64.DegToRad(fovX)/2)/aspectRatio)),
	}
}

func (r RenderContext) Width() int {
	return r.width
}
func (r RenderContext) Height() int {
	return r.height
}
func (r RenderContext) AspectRatio() float64 {
	return r.aspectRatio
}
func (r RenderContext) FovX() float64 {
	return r.fovX
}
func (r RenderContext) FovY() float64 {
	return r.fovY
}
