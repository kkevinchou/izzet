package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

const epsilon = 1e-9

func normalizeQuat(q mgl64.Quat) mgl64.Quat {
	if !finiteQuat(q) || q.Len() <= epsilon {
		return mgl64.QuatIdent()
	}
	return q.Normalize()
}

func safeNormalize(v, fallback mgl64.Vec3) mgl64.Vec3 {
	if !finiteVec3(v) || v.LenSqr() <= epsilon {
		if !finiteVec3(fallback) || fallback.LenSqr() <= epsilon {
			return mgl64.Vec3{0, 1, 0}
		}
		return fallback
	}
	return v.Normalize()
}

func finiteFloat(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func finiteVec3(v mgl64.Vec3) bool {
	return finiteFloat(v.X()) && finiteFloat(v.Y()) && finiteFloat(v.Z())
}

func finiteQuat(q mgl64.Quat) bool {
	return finiteFloat(q.W) && finiteVec3(q.V)
}

func componentMul(a, b mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z()}
}

func componentAbs(v mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{math.Abs(v.X()), math.Abs(v.Y()), math.Abs(v.Z())}
}

func clamp(value, minValue, maxValue float64) float64 {
	if !finiteFloat(value) {
		return minValue
	}
	return math.Max(minValue, math.Min(maxValue, value))
}

func clamp01(value float64) float64 {
	return clamp(value, 0, 1)
}
