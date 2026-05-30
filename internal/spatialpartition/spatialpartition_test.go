package spatialpartition_test

import (
	"testing"

	"github.com/kkevinchou/izzet/internal/spatialpartition"
)

func TestPartition(t *testing.T) {
	partitionCount := 3
	p := spatialpartition.NewSpatialPartition(5, partitionCount)
	expectedPartitionCount := partitionCount * partitionCount * partitionCount

	actualPartitionCount := 0
	for i := range p.Partitions {
		for j := range p.Partitions[i] {
			for range p.Partitions[i][j] {
				actualPartitionCount += 1
			}
		}
	}

	if actualPartitionCount != expectedPartitionCount {
		t.Errorf("expected %d partitions, but got %d", expectedPartitionCount, actualPartitionCount)
	}
}
