package gizmo

import "github.com/go-gl/mathgl/mgl64"

var T *TranslationGizmo

type AxisType string

var AxisTypeX AxisType = "X"
var AxisTypeY AxisType = "Y"
var AxisTypeZ AxisType = "Z"

func init() {
	T = &TranslationGizmo{active: true}
}

type TranslationGizmo struct {
	active bool

	xDelta float64
	yDelta float64
	zDelta float64
}

func (g *TranslationGizmo) Active() bool {
	return g.active
}

func (g *TranslationGizmo) Reset() {
	g.active = false
	g.xDelta = 0
	g.yDelta = 0
	g.zDelta = 0
}

func (g *TranslationGizmo) Move(axisType AxisType, delta float64) bool {
	g.active = true

	if axisType == AxisTypeX {
		g.xDelta += delta
	} else if axisType == AxisTypeY {
		g.yDelta += delta
	} else if axisType == AxisTypeZ {
		g.zDelta += delta
	}

	return true
}

func (g *TranslationGizmo) Translation() mgl64.Vec3 {
	return mgl64.Vec3{g.xDelta, g.yDelta, g.zDelta}
}
