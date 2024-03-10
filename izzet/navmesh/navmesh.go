package navmesh

import "github.com/kkevinchou/kitolib/collision/collider"

type NavigationMesh2 struct {
	HeightField *HeightField
	Volume      collider.BoundingBox
}
