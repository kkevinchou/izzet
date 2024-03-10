package navmesh

import (
	"cmp"

	"github.com/kkevinchou/kitolib/collision/collider"
)

const maxHeight int = 0xffff

type NavigationMesh struct {
	HeightField *HeightField
	Volume      collider.BoundingBox
}

func Clamp[T cmp.Ordered](value, minInclusive, maxInclusive T) T {
	if value < minInclusive {
		return minInclusive
	} else if value > maxInclusive {
		return maxInclusive
	}
	return value
}
