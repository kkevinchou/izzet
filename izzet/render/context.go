package render

import "github.com/go-gl/mathgl/mgl64"

type ViewerContext struct {
	Position    mgl64.Vec3
	Orientation mgl64.Quat

	InverseViewMatrix mgl64.Mat4
	ProjectionMatrix  mgl64.Mat4
}

type LightContext struct {
	LightSpaceMatrix mgl64.Mat4
}
