package navmesh

import (
	"cmp"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/collision/collider"
)

const maxHeight int = 0xffff
const maxDistance int = 0xffff

// / Defines the number of bits allocated to rcSpan::smin and rcSpan::smax.
const spanHeightBits int = 13

// / Defines the maximum value for rcSpan::smin and rcSpan::smax.
const spanMaxHeight int = (1 << spanHeightBits) - 1

type AREA_TYPE int

const NULL_AREA AREA_TYPE = 0
const WALKABLE_AREA AREA_TYPE = 63

// match what recast uses, they also reserve 63 as unconnected but we don't use that
const maxLayers int = 62

var dirs [4]int = [4]int{0, 1, 2, 3}

// // CW ordering
// var xDirs [4]int = [4]int{-1, 0, 1, 0}
// var zDirs [4]int = [4]int{0, -1, 0, 1}

// const dirLeft int = 0
// const dirUp int = 1
// const dirRight int = 2
// const dirDown int = 3

// CCW ordering
var xDirs [4]int = [4]int{-1, 0, 1, 0}
var zDirs [4]int = [4]int{0, 1, 0, -1}

const dirLeft int = 0
const dirUp int = 3
const dirRight int = 2
const dirDown int = 1

type NavigationMesh struct {
	Invalidated          bool
	InvalidatedTimestamp int

	HeightField        *HeightField
	CompactHeightField *CompactHeightField
	Volume             collider.BoundingBox
	BlurredDistances   []int
	DebugLines         [][2]mgl64.Vec3
	ContourSet         *ContourSet
	DetailedMesh       *DetailedMesh
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
