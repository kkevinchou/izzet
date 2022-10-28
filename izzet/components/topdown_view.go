package components

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	utils "github.com/kkevinchou/izzet/lib/libutils"
)

const (
	cameraRotationXMax = 80
	cameraSpeedScalar  = 10
	xSensitivity       = float64(0.2)
	ySensitivity       = float64(0.1)
)

type TopDownViewComponent struct {
	view mgl64.Vec2
}

func (c *TopDownViewComponent) View() mgl64.Vec2 {
	return c.view
}

func (c *TopDownViewComponent) SetView(view mgl64.Vec2) {
	c.view = view
}

func (c *TopDownViewComponent) UpdateView(delta mgl64.Vec2) {
	c.view[1] += delta.X() * xSensitivity
	c.view[0] += delta.Y() * ySensitivity
	// if c.view.Y > 360 {
	// 	c.view.Y -= 360
	// }
	// if c.view.Y < -360 {
	// 	c.view.Y += 360
	// }
}

func (c *TopDownViewComponent) Forward() mgl64.Vec3 {
	xRadianAngle := -toRadians(c.view.X())
	if xRadianAngle < 0 {
		xRadianAngle += 2 * math.Pi
	}

	yRadianAngle := -(toRadians(0) - (math.Pi / 2))
	if yRadianAngle < 0 {
		yRadianAngle += 2 * math.Pi
	}

	x := math.Cos(yRadianAngle) * math.Cos(xRadianAngle)
	y := math.Sin(xRadianAngle)
	z := -math.Sin(yRadianAngle) * math.Cos(xRadianAngle)

	return mgl64.Vec3{x, y, z}.Mul(-1)
}

func (c *TopDownViewComponent) Right() mgl64.Vec3 {
	xRadianAngle := -toRadians(c.view.X())
	if xRadianAngle < 0 {
		xRadianAngle += 2 * math.Pi
	}
	yRadianAngle := -(toRadians(0) - (math.Pi / 2))
	if yRadianAngle < 0 {
		yRadianAngle += 2 * math.Pi
	}

	x, y, z := math.Cos(yRadianAngle), math.Sin(xRadianAngle), -math.Sin(yRadianAngle)

	v1 := mgl64.Vec3{x, math.Abs(y), z}
	v2 := mgl64.Vec3{x, 0, z}
	v3 := v1.Cross(v2)

	if utils.Vec3IsZero(v3) {
		v3 = mgl64.Vec3{v2.Z(), 0, -v2.X()}
	}

	return v3.Normalize()
}

func toRadians(degrees float64) float64 {
	return degrees / 180 * math.Pi
}

func (c *TopDownViewComponent) AddToComponentContainer(container *ComponentContainer) {
	container.TopDownViewComponent = c
}
