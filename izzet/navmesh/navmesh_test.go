package navmesh_test

import (
	"testing"

	"github.com/kkevinchou/izzet/izzet/navmesh"
)

func TestMain(t *testing.T) {
	navmesh.Plinex(0, 0, 0, 50, 50, 0)

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
