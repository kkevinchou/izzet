package navmesh

import (
	"cmp"

	"github.com/kkevinchou/kitolib/collision/collider"
)

const maxHeight int = 0xffff
const maxDistance int = 0xffff

// match what recast uses, they also reserve 63 as unconnected but we don't use that
const maxLayers int = 62

var dirs [4]int = [4]int{0, 1, 2, 3}
var xDirs [4]int = [4]int{-1, 0, 1, 0}
var zDirs [4]int = [4]int{0, 1, 0, -1}

const dirLeft int = 0
const dirDown int = 1
const dirRight int = 2
const dirUp int = 3

type NavigationMesh struct {
	HeightField *HeightField
	Volume      collider.BoundingBox
}

func Min[T cmp.Ordered](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T cmp.Ordered](a T, b T) T {
	if a > b {
		return a
	}
	return b
}

func Clamp[T cmp.Ordered](value, minInclusive, maxInclusive T) T {
	if value < minInclusive {
		return minInclusive
	} else if value > maxInclusive {
		return maxInclusive
	}
	return value
}

func Abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
