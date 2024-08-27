package noise

import "math/rand"

// 3d worley noise
func Worley3D(width, height, depth int) []float32 {
	result := make([]float32, width*height*depth*3)

	for z := range depth {
		for y := range height {
			for x := range width {
				result[(x+(width*y)+(width*height*z))*3] = float32(x) + rand.Float32()
				result[(x+(width*y)+(width*height*z))*3+1] = float32(y) + rand.Float32()
				result[(x+(width*y)+(width*height*z))*3+2] = float32(z) + rand.Float32()
			}
		}
	}

	return result
}

// x+(xMax*y)+(xMax*yMax*z)
