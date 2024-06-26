package entities

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
)

type PositionSync struct {
	Goal                   mgl64.Vec3
	Active                 bool
	StartTime              time.Time
	CorrectionVelocity     mgl64.Vec3
	CorrectionAcceleration mgl64.Vec3
}
