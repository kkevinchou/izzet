package gizmo

import "github.com/go-gl/mathgl/mgl64"

var T *TranslationGizmo

type TranslationGizmo struct {
	TranslationDir     mgl64.Vec3
	MotionPivot        mgl64.Vec3
	Active             bool
	Axes               []mgl64.Vec3
	HoverIndex         int
	ActivationPosition mgl64.Vec3
}
