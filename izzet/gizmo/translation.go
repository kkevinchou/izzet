package gizmo

import "github.com/go-gl/mathgl/mgl64"

var T *TranslationGizmo

type TranslationGizmo struct {
	TranslationDir     mgl64.Vec3
	OldClosestPoint    mgl64.Vec3
	OldWorldPosition   mgl64.Vec3
	Active             bool
	Axes               []mgl64.Vec3
	HoverIndex         int
	ActivationPosition mgl64.Vec3
}

func (g *TranslationGizmo) Reset() {
	g.HoverIndex = -1
	g.ActivationPosition = mgl64.Vec3{}
	g.Active = false
}
