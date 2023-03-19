package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

// var NavMesh *NavigationMesh = New()

type World interface {
	SpatialPartition() *spatialpartition.SpatialPartition
	GetEntityByID(id int) *entities.Entity
}

type NavigationMesh struct {
	Volume collider.BoundingBox
	world  World
}

func New(world World) *NavigationMesh {
	return &NavigationMesh{
		Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-200, -200, -200}, MaxVertex: mgl64.Vec3{200, 200, 200}},
		world:  world,
	}
}

func (n *NavigationMesh) BakeNavMesh() {
	spatialPartition := n.world.SpatialPartition()
	entities := spatialPartition.QueryEntities(n.Volume)
	for _, entity := range entities {
		e := n.world.GetEntityByID(entity.GetID())
		_ = e
	}
}
