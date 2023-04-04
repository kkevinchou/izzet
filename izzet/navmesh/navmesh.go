package navmesh

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/spatialpartition"
)

// GENERAL IMPLEMENTATION NOTES

// in general the generated regions in a nav mesh should be free of holes.
// due to the resolution of the voxels there are degenerate cases where holes
// can be present in the generated mesh regions. for example, holes in meshes
// with a size of 1 or 2 tend to be ignored. however, larger holes will be properly
// processed

const stepHeight int = 4
const agentHeight int = 30

type World interface {
	SpatialPartition() *spatialpartition.SpatialPartition
	GetEntityByID(id int) *entities.Entity
}

type NavigationMesh struct {
	Volume collider.BoundingBox
	world  World

	voxelCount     int
	voxelField     [][][]Voxel
	voxelDimension float64
}

func New(world World) *NavigationMesh {
	nm := &NavigationMesh{
		// Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{75, -50, -200}, MaxVertex: mgl64.Vec3{350, 25, -50}},
		Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -50, -350}, MaxVertex: mgl64.Vec3{350, 150, 150}},
		// Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -25, -150}, MaxVertex: mgl64.Vec3{150, 150, 0}},
		// Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{0, -25, 0}, MaxVertex: mgl64.Vec3{100, 100, 150}},
		// Volume:         collider.BoundingBox{MinVertex: mgl64.Vec3{-50, -25, 0}, MaxVertex: mgl64.Vec3{100, 100, 150}},
		voxelDimension: 1.0,
		world:          world,
	}
	nm.BakeNavMesh()
	// move the scene out of the way
	entity := world.GetEntityByID(3)
	entities.SetLocalPosition(entity, mgl64.Vec3{0, -1000, 0})
	return nm
}

func (n *NavigationMesh) BakeNavMesh() {
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var dimensions [3]int = [3]int{int(delta[0] / n.voxelDimension), int(delta[1] / n.voxelDimension), int(delta[2] / n.voxelDimension)}

	n.voxelField = n.voxelize()
	buildNavigableArea(n.voxelField, dimensions)
	reachField := computeReachField(n.voxelField, dimensions)
	computeDistanceTransform(n.voxelField, reachField, dimensions)
	blurDistanceField(n.voxelField, reachField, dimensions)
	regionMap := watershed(n.voxelField, reachField, dimensions)
	_ = regionMap
	mergeRegions(n.voxelField, reachField, dimensions, regionMap)
}

type ReachInfo struct {
	sourceVoxel *Voxel
}

func computeReachField(voxelField [][][]Voxel, dimensions [3]int) [][][]ReachInfo {
	reachField := make([][][]ReachInfo, dimensions[0])
	for i := range reachField {
		reachField[i] = make([][]ReachInfo, dimensions[1])
		for j := range reachField[i] {
			reachField[i][j] = make([]ReachInfo, dimensions[2])
		}
	}

	for y := 0; y < dimensions[1]-stepHeight; y++ {
		for x := 0; x < dimensions[0]; x++ {
			for z := 0; z < dimensions[2]; z++ {
				if !voxelField[x][y][z].Filled {
					continue
				}

				for i := 1; i < stepHeight+1; i++ {
					if y+i < dimensions[1] {
						reachField[x][y+i][z].sourceVoxel = &voxelField[x][y][z]
					}

					if y-i >= 0 {
						reachField[x][y-i][z].sourceVoxel = &voxelField[x][y][z]
					}
				}
			}
		}
	}

	return reachField
}

var neighborDirs [][2]int = [][2]int{
	[2]int{-1, -1}, [2]int{0, -1}, [2]int{1, -1},
	[2]int{-1, 0} /* current voxel */, [2]int{1, 0},
	[2]int{-1, 1}, [2]int{0, 1}, [2]int{1, 1},
}

type VoxelPosition [3]int

func voxelPos(voxel *Voxel) VoxelPosition {
	return VoxelPosition{voxel.X, voxel.Y, voxel.Z}
}

func getNeighbors(x, y, z int, voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) []*Voxel {
	var neighbors []*Voxel

	for _, dir := range neighborDirs {
		if x+dir[0] < 0 || z+dir[1] < 0 || x+dir[0] >= dimensions[0] || z+dir[1] >= dimensions[2] {
			continue
		}

		var neighbor *Voxel
		reachNeighbor := &reachField[x+dir[0]][y][z+dir[1]]
		if reachNeighbor.sourceVoxel != nil {
			neighbor = reachNeighbor.sourceVoxel
		} else if voxelField[x+dir[0]][y][z+dir[1]].Filled {
			neighbor = &voxelField[x+dir[0]][y][z+dir[1]]
		}

		if neighbor == nil {
			continue
		}

		neighbors = append(neighbors, neighbor)
	}

	return neighbors
}
