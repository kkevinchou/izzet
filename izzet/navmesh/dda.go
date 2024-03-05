package navmesh

import "math"

func Plinex(x0, y0, z0, x1, y1, z1 int, LY, LZ, RY, RZ []int, vxs int) {
	var i, n, cx, cy, cz, sx, sy, sz int
	var by, bz []int

	// target buffer & order points by x
	if x1 >= x0 {
		i = x0
		x0 = x1
		x1 = i

		i = y0
		y0 = y1
		y1 = i

		i = z0
		z0 = z1
		z1 = i

		by = LY
		bz = LZ
	} else {
		by = RY
		bz = RZ
	}

	// line DDA parameters
	x1 -= x0
	sx = 0
	if x1 > 0 {
		sx = 1
	}
	if x1 < 0 {
		sx = -1
		x1 = -x1
	}
	if x1 != 0 {
		x1++
	}
	n = x1

	y1 -= y0
	sy = 0
	if y1 > 0 {
		sy = 1
	}
	if y1 < 0 {
		sy = -1
		y1 = -y1
	}
	if y1 != 0 {
		y1++
	}
	if n < y1 {
		n = y1
	}

	z1 -= z0
	sz = 0
	if z1 > 0 {
		sz = 1
	}
	if z1 < 0 {
		sz = -1
		z1 = -z1
	}
	if z1 != 0 {
		z1++
	}
	if n < z1 {
		n = z1
	}

	// single pixel (not a line)
	if n == 0 {
		if x0 >= 0 && x0 < len(LY) {
			LY[x0] = y0
			LZ[x0] = z0
			RY[x0] = y0
			RZ[x0] = z0
		}
		return
	}

	// ND DDA algo i is parameter
	for cx, cy, cz, i = n, n, n, 0; i < n; i++ {
		if x0 >= 0 && x0 < vxs {
			by[x0] = y0
			bz[x0] = z0
		}
		cx -= x1
		if cx <= 0 {
			cx += n
			x0 += sx
		}
		cy -= y1
		if cy <= 0 {
			cy += n
			y0 += sy
		}
		cz -= z1
		if cz <= 0 {
			cz += n
			z0 += sz
		}
	}
}

func Pliney(x0, y0, z0, x1, y1, z1 int, LX, LZ, RX, RZ []int, vys int) {
	var i, n, cx, cy, cz, sx, sy, sz int
	var bx, bz []int

	// target buffer & order points by y
	if y1 >= y0 {
		i = x0
		x0 = x1
		x1 = i

		i = y0
		y0 = y1
		y1 = i

		i = z0
		z0 = z1
		z1 = i

		bx = LX
		bz = LZ
	} else {
		bx = RX
		bz = RZ
	}

	// line DDA parameters
	x1 -= x0
	sx = 0
	if x1 > 0 {
		sx = 1
	}
	if x1 < 0 {
		sx = -1
		x1 = -x1
	}
	if x1 != 0 {
		x1++
	}
	n = x1

	y1 -= y0
	sy = 0
	if y1 > 0 {
		sy = 1
	}
	if y1 < 0 {
		sy = -1
		y1 = -y1
	}
	if y1 != 0 {
		y1++
	}
	if n < y1 {
		n = y1
	}

	z1 -= z0
	sz = 0
	if z1 > 0 {
		sz = 1
	}
	if z1 < 0 {
		sz = -1
		z1 = -z1
	}
	if z1 != 0 {
		z1++
	}
	if n < z1 {
		n = z1
	}

	// single pixel (not a line)
	if n == 0 {
		if y0 >= 0 && y0 < vys {
			LX[y0] = x0
			LZ[y0] = z0
			RX[y0] = x0
			RZ[y0] = z0
		}
		return
	}

	// ND DDA algo i is parameter
	for cx, cy, cz, i = n, n, n, 0; i < n; i++ {
		if y0 >= 0 && y0 < len(LX) {
			bx[y0] = x0
			bz[y0] = z0
		}
		cx -= x1
		if cx <= 0 {
			cx += n
			x0 += sx
		}
		cy -= y1
		if cy <= 0 {
			cy += n
			y0 += sy
		}
		cz -= z1
		if cz <= 0 {
			cz += n
			z0 += sz
		}
	}
}

func Plinez(x0, y0, z0, x1, y1, z1 int, LX, LY, RX, RY []int, vzs int) {
	var i, n, cx, cy, cz, sx, sy, sz int
	var bx, by []int

	// target buffer & order points by z
	if z1 >= z0 {
		i = x0
		x0 = x1
		x1 = i

		i = y0
		y0 = y1
		y1 = i

		i = z0
		z0 = z1
		z1 = i

		bx = LX
		by = LY
	} else {
		bx = RX
		by = RY
	}

	// line DDA parameters
	x1 -= x0
	sx = 0
	if x1 > 0 {
		sx = 1
	}
	if x1 < 0 {
		sx = -1
		x1 = -x1
	}
	if x1 != 0 {
		x1++
	}
	n = x1

	y1 -= y0
	sy = 0
	if y1 > 0 {
		sy = 1
	}
	if y1 < 0 {
		sy = -1
		y1 = -y1
	}
	if y1 != 0 {
		y1++
	}
	if n < y1 {
		n = y1
	}

	z1 -= z0
	sz = 0
	if z1 > 0 {
		sz = 1
	}
	if z1 < 0 {
		sz = -1
		z1 = -z1
	}
	if z1 != 0 {
		z1++
	}
	if n < z1 {
		n = z1
	}

	// single pixel (not a line)
	if n == 0 {
		if z0 >= 0 && z0 < vzs {
			LX[z0] = x0
			LY[z0] = y0
			RX[z0] = x0
			RY[z0] = y0
		}
		return
	}

	// ND DDA algo i is parameter
	for cx, cy, cz, i = n, n, n, 0; i < n; i++ {
		if z0 >= 0 && z0 < len(LX) {
			bx[z0] = x0
			by[z0] = y0
		}
		cx -= x1
		if cx <= 0 {
			cx += n
			x0 += sx
		}
		cy -= y1
		if cy <= 0 {
			cy += n
			y0 += sy
		}
		cz -= z1
		if cz <= 0 {
			cz += n
			z0 += sz
		}
	}
}

func Line(x0, y0, z0, x1, y1, z1 int, c float32, vxs, vys, vzs int, map3D [][][]float32) {
	var i, n, cx, cy, cz, sx, sy, sz int

	// line DDA parameters
	x1 -= x0
	sx = 0
	if x1 > 0 {
		sx = 1
	}
	if x1 < 0 {
		sx = -1
		x1 = -x1
	}
	if x1 != 0 {
		x1++
	}
	n = x1

	y1 -= y0
	sy = 0
	if y1 > 0 {
		sy = 1
	}
	if y1 < 0 {
		sy = -1
		y1 = -y1
	}
	if y1 != 0 {
		y1++
	}
	if n < y1 {
		n = y1
	}

	z1 -= z0
	sz = 0
	if z1 > 0 {
		sz = 1
	}
	if z1 < 0 {
		sz = -1
		z1 = -z1
	}
	if z1 != 0 {
		z1++
	}
	if n < z1 {
		n = z1
	}

	// single pixel (not a line)
	if n == 0 {
		if x0 >= 0 && x0 < vxs && y0 >= 0 && y0 < vys && z0 >= 0 && z0 < vzs {
			map3D[x0][y0][z0] = c
		}
		return
	}

	// ND DDA algo i is parameter
	for cx, cy, cz, i = n, n, n, 0; i < n; i++ {
		if x0 >= 0 && x0 < vxs && y0 >= 0 && y0 < vys && z0 >= 0 && z0 < vzs {
			map3D[x0][y0][z0] = c
		}
		cx -= x1
		if cx <= 0 {
			cx += n
			x0 += sx
		}
		cy -= y1
		if cy <= 0 {
			cy += n
			y0 += sy
		}
		cz -= z1
		if cz <= 0 {
			cz += n
			z0 += sz
		}
	}
}

func RasterizeTriangle(x0, y0, z0, x1, y1, z1, x2, y2, z2 int, map3D [][][]float32) {
	vxs := len(map3D)
	vys := len(map3D[0])
	vzs := len(map3D[0][0])
	vsz := int(math.Max(math.Max(float64(vxs), float64(vys)), float64(vzs)))

	LX := make([]int, vsz)
	LY := make([]int, vsz)
	LZ := make([]int, vsz)
	RX := make([]int, vsz)
	RY := make([]int, vsz)
	RZ := make([]int, vsz)

	var X0, Y0, Z0, X1, Y1, Z1, dx, dy, dz, x, y, z int

	// BBOX
	X0 = x0
	X1 = x0
	if X0 > x1 {
		X0 = x1
	}
	if X1 < x1 {
		X1 = x1
	}
	if X0 > x2 {
		X0 = x2
	}
	if X1 < x2 {
		X1 = x2
	}
	Y0 = y0
	Y1 = y0
	if Y0 > y1 {
		Y0 = y1
	}
	if Y1 < y1 {
		Y1 = y1
	}
	if Y0 > y2 {
		Y0 = y2
	}
	if Y1 < y2 {
		Y1 = y2
	}
	Z0 = z0
	Z1 = z0
	if Z0 > z1 {
		Z0 = z1
	}
	if Z1 < z1 {
		Z1 = z1
	}
	if Z0 > z2 {
		Z0 = z2
	}
	if Z1 < z2 {
		Z1 = z2
	}

	dx = X1 - X0
	dy = Y1 - Y0
	dz = Z1 - Z0

	if dx >= dy && dx >= dz { // x is major axis
		// render circumference into left/right buffers
		Plinex(x0, y0, z0, x1, y1, z1, LY, LZ, RY, RZ, vxs)
		Plinex(x1, y1, z1, x2, y2, z2, LY, LZ, RY, RZ, vxs)
		Plinex(x2, y2, z2, x0, y0, z0, LY, LZ, RY, RZ, vxs)

		// fill the triangle
		if X0 < 0 {
			X0 = 0
		}
		if X1 >= vxs {
			X1 = vxs - 1
		}
		for x = X0; x <= X1; x++ {
			y0 = LY[x]
			z0 = LZ[x]
			y1 = RY[x]
			z1 = RZ[x]
			Line(x, y0, z0, x, y1, z1, 1, vxs, vys, vzs, map3D)
		}
	} else if dy >= dx && dy >= dz { // y is major axis
		// render circumference into left/right buffers
		Pliney(x0, y0, z0, x1, y1, z1, LX, LZ, RX, RZ, vys)
		Pliney(x1, y1, z1, x2, y2, z2, LX, LZ, RX, RZ, vys)
		Pliney(x2, y2, z2, x0, y0, z0, LX, LZ, RX, RZ, vys)

		// fill the triangle
		if Y0 < 0 {
			Y0 = 0
		}
		if Y1 >= vys {
			Y1 = vys - 1
		}
		for y = Y0; y <= Y1; y++ {
			x0 = LX[y]
			z0 = LZ[y]
			x1 = RX[y]
			z1 = RZ[y]
			Line(x0, y, z0, x1, y, z1, 1, vxs, vys, vzs, map3D)
		}
	} else if dz >= dx && dz >= dy { // z is major axis
		// render circumference into left/right buffers
		Plinez(x0, y0, z0, x1, y1, z1, LX, LY, RX, RY, vzs)
		Plinez(x1, y1, z1, x2, y2, z2, LX, LY, RX, RY, vzs)
		Plinez(x2, y2, z2, x0, y0, z0, LX, LY, RX, RY, vzs)

		// fill the triangle
		if Z0 < 0 {
			Z0 = 0
		}
		if Z1 >= vzs {
			Z1 = vzs - 1
		}
		for z = Z0; z <= Z1; z++ {
			x0 = LX[z]
			y0 = LY[z]
			x1 = RX[z]
			y1 = RY[z]
			Line(x0, y0, z, x1, y1, z, 1, vxs, vys, vzs, map3D)
		}
	}
}
