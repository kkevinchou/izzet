package spatialpartition

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/settings"
)

type Entity interface {
	GetID() int
	Position() mgl64.Vec3
	BoundingBox() collider.BoundingBox
}

type Partition struct {
	Key      PartitionKey
	AABB     collider.BoundingBox
	entities map[int]Entity
}

func (p *Partition) String() string {
	return fmt.Sprintf("Partition %v", p.Key)
}

type PartitionKey [3]int

type CachedPosition struct {
	Valid    bool
	Position mgl64.Vec3
}

type CachedPartition struct {
	Valid      bool
	Partitions map[PartitionKey]any
}

type SpatialPartition struct {
	Partitions         [][][]Partition
	PartitionDimension int
	PartitionCount     int

	entityPartitionCache [settings.MaxEntityCount]CachedPartition
	entityPositionCache  [settings.MaxEntityCount]CachedPosition

	// pool partition keys to avoid reallocating new arrays and resizing slices
	pooledPartitionKeys []PartitionKey
}

// NewSpatialPartition creates a spatial partition with the bottom at <0, 0, 0>
// the spatial partition spans the rectangular space for
// d = partitionDimension * partitionCount
// <-d, 0, -d> to <d, 2 * d, d>
func NewSpatialPartition(partitionDimension int, partitionCount int) *SpatialPartition {
	s := &SpatialPartition{
		PartitionDimension: partitionDimension,
		PartitionCount:     partitionCount,
	}

	s.initialize()

	return s
}

func (s *SpatialPartition) initialize() {
	s.Partitions = initializePartitions(s.PartitionDimension, s.PartitionCount)
	s.pooledPartitionKeys = make([]PartitionKey, s.PartitionCount*s.PartitionCount*s.PartitionCount)
	s.entityPartitionCache = [settings.MaxEntityCount]CachedPartition{}
	s.entityPositionCache = [settings.MaxEntityCount]CachedPosition{}
}

func (s *SpatialPartition) Clear() {
	s.initialize()
}

// QueryEntities queries for entities that exist in partitions that the boundingBox is a part of
func (s *SpatialPartition) QueryEntities(boundingBox collider.BoundingBox) []Entity {
	// determine which partitions the entity touches
	// collect all entities that belong to each of the partitions

	partitions := s.IntersectingPartitions(boundingBox)
	candidates := []Entity{}

	var seenEntities [settings.MaxEntityCount]bool

	for _, partitionKey := range partitions {
		partition := &s.Partitions[partitionKey[0]][partitionKey[1]][partitionKey[2]]
		for _, e := range partition.entities {
			if !seenEntities[e.GetID()] {
				seenEntities[e.GetID()] = true
				candidates = append(candidates, e)
			}
		}
	}

	return candidates
}

func (s *SpatialPartition) IndexEntities(entityList []Entity) {
	for _, entity := range entityList {
		if s.entityPositionCache[entity.GetID()].Valid {
			if s.entityPositionCache[entity.GetID()].Position == entity.Position() {
				continue
			}
		}
		s.entityPositionCache[entity.GetID()].Position = entity.Position()
		s.entityPositionCache[entity.GetID()].Valid = true

		// remove from old partitions
		oldPartitions := s.entityPartitionCache[entity.GetID()].Partitions
		for partitionKey := range oldPartitions {
			partition := &s.Partitions[partitionKey[0]][partitionKey[1]][partitionKey[2]]
			delete(partition.entities, entity.GetID())
		}

		// add to new partitions
		boundingBox := entity.BoundingBox()
		newPartitions := s.IntersectingPartitions(boundingBox)
		for _, partitionKey := range newPartitions {
			partition := &s.Partitions[partitionKey[0]][partitionKey[1]][partitionKey[2]]
			partition.entities[entity.GetID()] = entity
			if !s.entityPartitionCache[entity.GetID()].Valid {
				s.entityPartitionCache[entity.GetID()].Partitions = map[PartitionKey]any{}
				s.entityPartitionCache[entity.GetID()].Valid = true
			}
			s.entityPartitionCache[entity.GetID()].Partitions[partition.Key] = partitionKey
		}
	}
}

func initializePartitions(partitionDimension int, partitionCount int) [][][]Partition {
	d := partitionDimension * partitionCount
	partitions := make([][][]Partition, partitionCount)
	for i := 0; i < partitionCount; i++ {
		partitions[i] = make([][]Partition, partitionCount)
		for j := 0; j < partitionCount; j++ {
			partitions[i][j] = make([]Partition, partitionCount)
			for k := 0; k < partitionCount; k++ {
				partitions[i][j][k] = Partition{
					Key: PartitionKey{i, j, k},
					AABB: collider.BoundingBox{
						MinVertex: mgl64.Vec3{float64(i*partitionDimension - d/2), float64(j*partitionDimension - d/2), float64(k*partitionDimension - d/2)},
						MaxVertex: mgl64.Vec3{float64((i+1)*partitionDimension - d/2), float64((j+1)*partitionDimension - d/2), float64((k+1)*partitionDimension - d/2)},
					},
					entities: map[int]Entity{},
				}
			}
		}
	}
	return partitions
}

func (s *SpatialPartition) IntersectingPartitions(boundingBox collider.BoundingBox) []PartitionKey {
	i1, j1, k1, found1 := s.VertexToPartitionClamped(boundingBox.MinVertex, true, false)
	if !found1 {
		return nil
	}

	i2, j2, k2, found2 := s.VertexToPartitionClamped(boundingBox.MaxVertex, false, true)
	if !found2 {
		return nil
	}

	index := -1
	for i := 0; i <= i2-i1; i++ {
		for j := 0; j <= j2-j1; j++ {
			for k := 0; k <= k2-k1; k++ {
				index += 1
				s.pooledPartitionKeys[index] = PartitionKey{i1 + i, j1 + j, k1 + k}
			}
		}
	}

	if index == -1 {
		return nil
	}

	return s.pooledPartitionKeys[0 : index+1]
}

func (s *SpatialPartition) calcMinPartitionVertex() mgl64.Vec3 {
	d := s.PartitionDimension * s.PartitionCount
	return mgl64.Vec3{float64(-d / 2), float64(-d / 2), float64(-d / 2)}
}

func (s *SpatialPartition) calcMaxPartitionVertex() mgl64.Vec3 {
	d := s.PartitionDimension * s.PartitionCount
	return mgl64.Vec3{float64((s.PartitionCount)*s.PartitionDimension - d/2), float64((s.PartitionCount)*s.PartitionDimension - d/2), float64((s.PartitionCount)*s.PartitionDimension - d/2)}
}

func (s *SpatialPartition) VertexToPartitionClamped(vertex mgl64.Vec3, clampMin, clampMax bool) (int, int, int, bool) {
	var i, j, k int
	var clampI, clampJ, clampK bool

	minPartitionVertex := s.calcMinPartitionVertex()
	maxPartitionVertex := s.calcMaxPartitionVertex()

	minDelta := vertex.Sub(minPartitionVertex)
	if clampMin {
		if minDelta.X() < 0 {
			i = 0
			clampI = true
		}
		if minDelta.Y() < 0 {
			j = 0
			clampJ = true
		}
		if minDelta.Z() < 0 {
			k = 0
			clampK = true
		}
	} else {
		if minDelta.X() < 0 || minDelta.Y() < 0 || minDelta.Z() < 0 {
			return 0, 0, 0, false
		}
	}

	maxDelta := vertex.Sub(maxPartitionVertex)
	if clampMax {
		if maxDelta.X() >= 0 {
			i = s.PartitionCount - 1
			clampI = true
		}
		if maxDelta.Y() >= 0 {
			j = s.PartitionCount - 1
			clampJ = true
		}
		if maxDelta.Z() >= 0 {
			k = s.PartitionCount - 1
			clampK = true
		}
	} else {
		if maxDelta.X() > 0 || maxDelta.Y() > 0 || maxDelta.Z() > 0 {
			return 0, 0, 0, false
		}
	}

	if !clampI {
		i = int(minDelta.X() / float64(s.PartitionDimension))
	}

	if !clampJ {
		j = int(minDelta.Y() / float64(s.PartitionDimension))
	}

	if !clampK {
		k = int(minDelta.Z() / float64(s.PartitionDimension))
	}

	return i, j, k, true
}

func (s *SpatialPartition) DeleteEntity(entityID int) {
	partitions := s.entityPartitionCache[entityID].Partitions
	for partitionKey := range partitions {
		delete(s.Partitions[partitionKey[0]][partitionKey[1]][partitionKey[2]].entities, entityID)
	}
	s.entityPartitionCache[entityID].Valid = false
	s.entityPositionCache[entityID].Valid = false
}
