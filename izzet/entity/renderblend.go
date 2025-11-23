package entity

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
)

type RenderBlend struct {
	StartTime          time.Time
	StartFrame         int
	Active             bool
	BlendStartPosition mgl64.Vec3
}
