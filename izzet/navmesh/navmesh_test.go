package navmesh_test

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	// navmesh.Plinex2(0, 0, 0, 50, 50, 0)
	// fmt.Println(navmesh.LY)
	// fmt.Println(navmesh.RY)
	// fmt.Println(navmesh.LZ)
	// fmt.Println(navmesh.RZ)

	// var map3D [navmesh.BufferDimension][navmesh.BufferDimension][navmesh.BufferDimension]float32
	// navmesh.TriangleComp(0, 0, 0, 45, 0, 0, 45, 45, 0, 1, navmesh.BufferDimension, navmesh.BufferDimension, navmesh.BufferDimension, &map3D)
	// fmt.Println(map3D)
	// for i := navmesh.BufferDimension - 1; i >= 0; i-- {
	// 	for j := 0; j < navmesh.BufferDimension; j++ {
	// 		fmt.Print(map3D[j][i][0])
	// 	}
	// 	fmt.Printf("\n")
	// }
}

func TestMain2(t *testing.T) {
	F()
	// memory usage is still 15GB
	fmt.Println("HI")
}

type Voxel2 struct {
	// Filled bool
	X int
	Y int
	Z int
}

func F() {
	vxs := 1000
	vys := 500
	vzs := 1000

	// var map3D [1000][500][1000]Voxel2
	// _ = map3D
	map3D := make([][][]Voxel2, vxs)
	for x := range vxs {
		map3D[x] = make([][]Voxel2, vys)
		for y := range vys {
			map3D[x][y] = make([]Voxel2, vzs)
		}
	}

	for i := 0; i < vxs; i++ {
		for j := 0; j < vys; j++ {
			for k := 0; k < vzs; k++ {
				_ = i
				_ = j
				_ = k

				// memory usage increases by 15GB
				// voxel := map3D[i][j][k]
				// _ = voxel
			}
		}
	}
	fmt.Println("HI")
}
