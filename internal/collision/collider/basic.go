package collider

import "github.com/go-gl/mathgl/mgl64"

type Plane struct {
	Point  mgl64.Vec3
	Normal mgl64.Vec3
}

type Ray struct {
	Origin    mgl64.Vec3
	Direction mgl64.Vec3
}

type Line struct {
	P1 mgl64.Vec3
	P2 mgl64.Vec3
}
