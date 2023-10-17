package gizmo

import "github.com/go-gl/mathgl/mgl64"

var S *ScaleGizmo

type AxisType int

const (
	NullAxis AxisType = iota
	AllAxis
	XAxis
	YAxis
	ZAxis
)

type Axis struct {
	Vector mgl64.Vec3
	Type   AxisType
}

type ScaleGizmo struct {
	OldMousePosition mgl64.Vec2
	Active           bool
	Axes             []Axis
	HoveredAxisType  AxisType
	ActivationScale  mgl64.Vec3
}

func (g *ScaleGizmo) Reset() {
	g.HoveredAxisType = NullAxis
	g.ActivationScale = mgl64.Vec3{}
	g.Active = false
}
