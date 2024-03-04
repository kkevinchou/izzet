package navmesh

const BufferDimension = 50
const vxs, vys, vzs int = BufferDimension, BufferDimension, BufferDimension

var lx [BufferDimension]int
var ly [BufferDimension]int
var lz [BufferDimension]int
var rx [BufferDimension]int
var ry [BufferDimension]int
var rz [BufferDimension]int

func Plinex(x0, y0, z0, x1, y1, z1 int) {
	var i, n, cx, cy, cz, sx, sy, sz int
	var by, bz *[BufferDimension]int

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

		by = &ly
		bz = &lz
	} else {
		by = &ry
		bz = &rz
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
	if y1 != 0 && n < y1 {
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
	if z1 != 0 && n < z1 {
		n = z1
	}

	// single pixel (not a line)
	if n == 0 {
		if x0 >= 0 && x0 < len(ly) {
			ly[x0] = y0
			lz[x0] = z0
			ry[x0] = y0
			rz[x0] = z0
		}
		return
	}

	// ND DDA algo i is parameter
	for cx, cy, cz, i = n, n, n, 0; i < n; i++ {
		if x0 >= 0 && x0 < len(ly) {
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

func Pliney(x0, y0, z0, x1, y1, z1 int) {
	var i, n, cx, cy, cz, sx, sy, sz int
	var bx, bz *[BufferDimension]int

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

		bx = &lx
		bz = &lz
	} else {
		bx = &rx
		bz = &rz
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
	if y1 != 0 && n < y1 {
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
	if z1 != 0 && n < z1 {
		n = z1
	}

	// single pixel (not a line)
	if n == 0 {
		if y0 >= 0 && y0 < len(lx) {
			lx[y0] = x0
			lz[y0] = z0
			rx[y0] = x0
			rz[y0] = z0
		}
		return
	}

	// ND DDA algo i is parameter
	for cx, cy, cz, i = n, n, n, 0; i < n; i++ {
		if y0 >= 0 && y0 < len(lx) {
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

func Plinez(x0, y0, z0, x1, y1, z1 int) {
	var i, n, cx, cy, cz, sx, sy, sz int
	var bx, by *[BufferDimension]int

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

		bx = &lx
		by = &ly
	} else {
		bx = &rx
		by = &ry
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
	if y1 != 0 && n < y1 {
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
	if z1 != 0 && n < z1 {
		n = z1
	}

	// single pixel (not a line)
	if n == 0 {
		if z0 >= 0 && z0 < len(lx) {
			lx[z0] = x0
			ly[z0] = y0
			rx[z0] = x0
			ry[z0] = y0
		}
		return
	}

	// ND DDA algo i is parameter
	for cx, cy, cz, i = n, n, n, 0; i < n; i++ {
		if z0 >= 0 && z0 < len(lx) {
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

func Line(x0, y0, z0, x1, y1, z1 int, c float32, vxs, vys, vzs int, map3D *[BufferDimension][BufferDimension][BufferDimension]float32) {
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
	if y1 != 0 && n < y1 {
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
	if z1 != 0 && n < z1 {
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

func TriangleComp(x0, y0, z0, x1, y1, z1, x2, y2, z2 int, c float32, vxs, vys, vzs int, map3D *[BufferDimension][BufferDimension][BufferDimension]float32) {
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
		Plinex(x0, y0, z0, x1, y1, z1)
		Plinex(x1, y1, z1, x2, y2, z2)
		Plinex(x2, y2, z2, x0, y0, z0)

		// fill the triangle
		if X0 < 0 {
			X0 = 0
		}
		if X1 >= vxs {
			X1 = vxs - 1
		}
		for x = X0; x <= X1; x++ {
			y0 = ly[x]
			z0 = lz[x]
			y1 = ry[x]
			z1 = rz[x]
			Line(x, y0, z0, x, y1, z1, c, vxs, vys, vzs, map3D)
		}
	} else if dy >= dx && dy >= dz { // y is major axis
		// render circumference into left/right buffers
		Pliney(x0, y0, z0, x1, y1, z1)
		Pliney(x1, y1, z1, x2, y2, z2)
		Pliney(x2, y2, z2, x0, y0, z0)

		// fill the triangle
		if Y0 < 0 {
			Y0 = 0
		}
		if Y1 >= vys {
			Y1 = vys - 1
		}
		for y = Y0; y <= Y1; y++ {
			x0 = lx[y]
			z0 = lz[y]
			x1 = rx[y]
			z1 = rz[y]
			Line(x0, y, z0, x1, y, z1, c, vxs, vys, vzs, map3D)
		}
	} else if dz >= dx && dz >= dy { // z is major axis
		// render circumference into left/right buffers
		Plinez(x0, y0, z0, x1, y1, z1)
		Plinez(x1, y1, z1, x2, y2, z2)
		Plinez(x2, y2, z2, x0, y0, z0)

		// fill the triangle
		if Z0 < 0 {
			Z0 = 0
		}
		if Z1 >= vzs {
			Z1 = vzs - 1
		}
		for z = Z0; z <= Z1; z++ {
			x0 = lx[z]
			y0 = ly[z]
			x1 = rx[z]
			y1 = ry[z]
			Line(x0, y0, z, x1, y1, z, c, vxs, vys, vzs, map3D)
		}
	}
}
