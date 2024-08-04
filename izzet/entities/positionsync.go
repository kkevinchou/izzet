package entities

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

type PositionSync struct {
	Goal                   mgl32.Vec3
	Active                 bool
	StartTime              time.Time
	CorrectionVelocity     mgl32.Vec3
	CorrectionAcceleration mgl32.Vec3
}
