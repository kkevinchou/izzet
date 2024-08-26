package noise

import "math/rand"

// 3d worley noise
func Worley3D(x, y, z int) []float32 {
	result := make([]float32, x*y*z*3)

	xMax, yMax, zMax := x, y, z

	for x := range xMax {
		for y := range yMax {
			for z := range zMax {
				result[x+(xMax*y)+(xMax*yMax*z)] = float32(x) + rand.Float32()
				result[x+(xMax*y)+(xMax*yMax*z)+1] = float32(y) + rand.Float32()
				result[x+(xMax*y)+(xMax*yMax*z)+2] = float32(z) + rand.Float32()
			}
		}
	}

	return result
}

// x+(xMax*y)+(xMax*yMax*z)
