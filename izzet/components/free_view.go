package components

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

type FreeViewComponent struct {
	view mgl64.Vec2
}

func (c *FreeViewComponent) View() mgl64.Vec2 {
	return c.view
}

func (c *FreeViewComponent) SetView(view mgl64.Vec2) {
	c.view = view
}

func (c *FreeViewComponent) UpdateView(delta mgl64.Vec2) {
	c.view[0] += delta.Y() * ySensitivity
	c.view[1] += delta.X() * xSensitivity

	if c.view.X() < -cameraRotationXMax {
		c.view[0] = -cameraRotationXMax
	} else if c.view.X() > cameraRotationXMax {
		c.view[0] = cameraRotationXMax
	}
}

func (c *FreeViewComponent) Forward() mgl64.Vec3 {
	xRadianAngle := -toRadians(c.view.X())
	if xRadianAngle < 0 {
		xRadianAngle += 2 * math.Pi
	}
	yRadianAngle := -(toRadians(c.view.Y()) - (math.Pi / 2))
	if yRadianAngle < 0 {
		yRadianAngle += 2 * math.Pi
	}

	x := math.Cos(yRadianAngle) * math.Cos(xRadianAngle)
	y := math.Sin(xRadianAngle)
	z := -math.Sin(yRadianAngle) * math.Cos(xRadianAngle)
	return mgl64.Vec3{x, y, z}.Mul(-1)
}

func (c *FreeViewComponent) Right() mgl64.Vec3 {
	xRadianAngle := -toRadians(c.view.X())
	if xRadianAngle < 0 {
		xRadianAngle += 2 * math.Pi
	}
	yRadianAngle := -(toRadians(c.view.Y()) - (math.Pi / 2))
	if yRadianAngle < 0 {
		yRadianAngle += 2 * math.Pi
	}

	x, y, z := math.Cos(yRadianAngle), math.Sin(xRadianAngle), -math.Sin(yRadianAngle)

	v1 := mgl64.Vec3{x, math.Abs(y), z}
	v2 := mgl64.Vec3{x, 0, z}
	v3 := v1.Cross(v2)

	if v3.X() == 0 && v3.Y() == 0 && v3.Z() == 0 {
		v3 = mgl64.Vec3{v2.Z(), 0, -v2.X()}
	}

	return v3.Normalize()
}
