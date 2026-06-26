package physics

import "github.com/go-gl/mathgl/mgl64"

const epsilon = 1e-9

func safeNormalize(v, fallback mgl64.Vec3) mgl64.Vec3 {
	if v.LenSqr() <= epsilon {
		if fallback.LenSqr() <= epsilon {
			return mgl64.Vec3{0, 1, 0}
		}
		return fallback
	}
	return v.Normalize()
}

func componentMul(a, b mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{a.X() * b.X(), a.Y() * b.Y(), a.Z() * b.Z()}
}

func componentAbs(v mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{mgl64.Abs(v.X()), mgl64.Abs(v.Y()), mgl64.Abs(v.Z())}
}
