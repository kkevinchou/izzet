package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

const epsilon = 1e-9

func normalizeQuat(q mgl64.Quat) mgl64.Quat {
	if q.Len() <= epsilon {
		return mgl64.QuatIdent()
	}
	return q.Normalize()
}

func safeNormalize(v, fallback mgl64.Vec3) mgl64.Vec3 {
	if v.LenSqr() <= epsilon {
		return fallback
	}
	return v.Normalize()
}

func componentMul(a, b mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z()}
}

func componentAbs(v mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{math.Abs(v.X()), math.Abs(v.Y()), math.Abs(v.Z())}
}

func clamp(value, minValue, maxValue float64) float64 {
	return math.Max(minValue, math.Min(maxValue, value))
}

func clamp01(value float64) float64 {
	return clamp(value, 0, 1)
}
