package collider

import "github.com/go-gl/mathgl/mgl64"

type Sphere struct {
	Center        mgl64.Vec3
	Radius        float64
	RadiusSquared float64
}

func NewSphere(center mgl64.Vec3, radius float64) Sphere {
	return Sphere{
		Center:        center,
		Radius:        radius,
		RadiusSquared: radius * radius,
	}
}
