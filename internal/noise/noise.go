package noise

import "math/rand"

type Point struct {
	X, Y, Z float32
}

// 3d worley noise
func Worley3D(x, y, z int) []Point {
	result := make([]Point, x*y*z)

	xMax, yMax, zMax := x, y, z

	for x := range xMax {
		for y := range yMax {
			for z := range zMax {
				result[x+(xMax*y)+(xMax*yMax*z)] = Point{
					X: float32(x) + rand.Float32(),
					Y: float32(y) + rand.Float32(),
					Z: float32(z) + rand.Float32(),
				}
			}
		}
	}

	return result
}

// x+(xMax*y)+(xMax*yMax*z)
