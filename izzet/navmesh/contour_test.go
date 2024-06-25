package navmesh

import (
	"fmt"
	"testing"
)

func TestArea2D(t *testing.T) {
	vertices := []SimplifiedVertex{
		{X: 100, Y: 0, Z: 0},
		{X: 100, Y: 0, Z: 100},
		{X: 0, Y: 0, Z: 100},
		{X: 0, Y: 0, Z: 0},
	}
	area := calcAreaOfPolygon2D(vertices)
	fmt.Println(area)
}
