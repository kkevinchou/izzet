package spatialpartition

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/components"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision/collider"
)

type World interface {
	// GetSingleton() *singleton.Singleton
	GetPlayerEntity() entities.Entity
	QueryEntity(componentFlags int) []entities.Entity
	// GetPlayer() *player.Player
	GetEntityByID(id int) entities.Entity
}

type Partition struct {
	x        int
	y        int
	z        int
	AABB     *collider.BoundingBox
	entities []entities.Entity
}

type SpatialPartition struct {
	world              World
	Partitions         [][][]*Partition
	PartitionDimension int
	PartitionCount     int
}

// NewSpatialPartition creates a spatial partition with the bottom at <0, 0, 0>
// the spatial partition spans the rectangular space for
// d = partitionDimension * partitionCount
// <-d, 0, -d> to <d, 2 * d, d>
func NewSpatialPartition(world World, partitionDimension int, partitionCount int) *SpatialPartition {
	return &SpatialPartition{
		world:              world,
		Partitions:         initializePartitions(partitionDimension, partitionCount),
		PartitionDimension: partitionDimension,
		PartitionCount:     partitionCount,
	}
}

// QueryCollisionCandidates queries for collision candidates that have been stored in
// the spatial partition
func (s *SpatialPartition) QueryCollisionCandidates(entity entities.Entity) []entities.Entity {
	// determine which partitions the entity touches
	// collect all entities that belong to each of the partitions

	cc := entity.GetComponentContainer()
	if cc.ColliderComponent.BoundingBoxCollider == nil {
		return nil
	}

	boundingBox := cc.ColliderComponent.BoundingBoxCollider.Transform(cc.TransformComponent.Position)
	partitions := s.intersectingPartitions(boundingBox)

	seen := map[int]bool{}
	candidates := []entities.Entity{}
	for _, p := range partitions {
		entities := p.entities
		for _, e := range entities {
			if _, ok := seen[e.GetID()]; !ok {
				seen[e.GetID()] = true
				candidates = append(candidates, e)
			}
		}
	}

	return candidates
}

func (s *SpatialPartition) AllCandidates() []entities.Entity {
	return s.world.QueryEntity(components.ComponentFlagCollider | components.ComponentFlagTransform)
}

func initializePartitions(partitionDimension int, partitionCount int) [][][]*Partition {
	d := partitionDimension * partitionCount
	partitions := make([][][]*Partition, partitionCount)
	for i := 0; i < partitionCount; i++ {
		partitions[i] = make([][]*Partition, partitionCount)
		for j := 0; j < partitionCount; j++ {
			partitions[i][j] = make([]*Partition, partitionCount)
			for k := 0; k < partitionCount; k++ {
				partitions[i][j][k] = &Partition{
					x: i,
					y: j,
					z: k,
					AABB: &collider.BoundingBox{
						MinVertex: mgl64.Vec3{float64(i*partitionDimension - d/2), float64(j*partitionDimension - d/2), float64(k*partitionDimension - d/2)},
						MaxVertex: mgl64.Vec3{float64((i+1)*partitionDimension - d/2), float64((j+1)*partitionDimension - d/2), float64((k+1)*partitionDimension - d/2)},
					},
				}
			}
		}
	}
	return partitions
}

func (s *SpatialPartition) FrameSetup(world World) {
	s.Partitions = initializePartitions(s.PartitionDimension, s.PartitionCount)
	entityList := world.QueryEntity(components.ComponentFlagCollider | components.ComponentFlagTransform)
	for _, entity := range entityList {
		cc := entity.GetComponentContainer()

		if cc.ColliderComponent.BoundingBoxCollider == nil {
			continue
		}

		boundingBox := cc.ColliderComponent.BoundingBoxCollider.Transform(cc.TransformComponent.Position)
		partitions := s.intersectingPartitions(boundingBox)
		// for i, p := range partitions {
		// 	fmt.Println("partition", i, ":", p.x, p.y, p.z)
		// }
		for _, p := range partitions {
			p.entities = append(p.entities, entity)
		}
	}
}

func (s *SpatialPartition) intersectingPartitions(boundingBox *collider.BoundingBox) []*Partition {
	// TODO(kchou): handle edge case where the vertex lies outside of the partitions. There may be some partitions
	// leading up to another vertex that we will not discover

	i1, j1, k1, found1 := s.VertexToPartition(boundingBox.MinVertex)
	if !found1 {
		return nil
	}

	i2, j2, k2, found2 := s.VertexToPartition(boundingBox.MaxVertex)
	if !found2 {
		return nil
	}

	partitions := []*Partition{}
	for i := 0; i <= i2-i1; i++ {
		for j := 0; j <= j2-j1; j++ {
			for k := 0; k <= k2-k1; k++ {
				partitions = append(partitions, s.Partitions[i1+i][j1+j][k1+k])
			}
		}
	}

	return partitions
}

func (s *SpatialPartition) VertexToPartition(vertex mgl64.Vec3) (int, int, int, bool) {
	d := s.PartitionDimension * s.PartitionCount
	minPartitionVertex := mgl64.Vec3{float64(-d / 2), float64(-d / 2), float64(-d / 2)}
	maxPartitionVertex := mgl64.Vec3{float64((s.PartitionCount)*s.PartitionDimension - d/2), float64((s.PartitionCount)*s.PartitionDimension - d/2), float64((s.PartitionCount)*s.PartitionDimension - d/2)}

	minDelta := vertex.Sub(minPartitionVertex)
	if minDelta.X() < 0 || minDelta.Y() < 0 || minDelta.Z() < 0 {
		return 0, 0, 0, false
	}

	maxDelta := vertex.Sub(maxPartitionVertex)
	if maxDelta.X() > 0 || maxDelta.Y() > 0 || maxDelta.Z() > 0 {
		return 0, 0, 0, false
	}

	i := int(minDelta.X() / float64(s.PartitionDimension))
	j := int(minDelta.Y() / float64(s.PartitionDimension))
	k := int(minDelta.Z() / float64(s.PartitionDimension))

	return i, j, k, true
}
