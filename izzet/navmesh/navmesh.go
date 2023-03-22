package navmesh

import (
	"fmt"
	"math"
	"sync"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/pathing"
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

	navmesh     *pathing.NavMesh
	clusterIDs  []int
	constructed bool

	vertices []mgl64.Vec3

	voxels []collider.BoundingBox
	mutex  sync.Mutex
}

func New(world World) *NavigationMesh {
	return &NavigationMesh{
		Volume: collider.BoundingBox{MinVertex: mgl64.Vec3{-150, -150, -150}, MaxVertex: mgl64.Vec3{150, 150, 150}},
		world:  world,
	}
}

type RenderData struct {
}

func (n *NavigationMesh) RenderData() ([]mgl64.Vec3, []int) {
	return n.vertices, n.clusterIDs
}

func (n *NavigationMesh) Voxels() []collider.BoundingBox {
	return n.voxels
}
func (n *NavigationMesh) Vertices() []mgl64.Vec3 {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.vertices
}

func (n *NavigationMesh) Voxelize() {
	spatialPartition := n.world.SpatialPartition()
	sEntities := spatialPartition.QueryEntities(n.Volume)

	if len(sEntities) == 0 {
		fmt.Println("not constructed")
		return
	}

	n.constructed = true
	var entities []*entities.Entity
	boundingBoxes := map[int]collider.BoundingBox{}
	entityTriCount := map[int]int{}

	for _, entity := range sEntities {
		e := n.world.GetEntityByID(entity.GetID())
		// if !strings.Contains(e.Name, "Tile") {
		// 	continue
		// }

		entities = append(entities, e)
		boundingBoxes[e.GetID()] = *e.BoundingBox()

		for _, rd := range e.Model.RenderData() {
			meshID := rd.MeshID
			numVerts := len(e.Model.Collection().Meshes[meshID].Vertices)
			entityTriCount[e.GetID()] = numVerts / 3
		}
	}

	outputWork := make(chan collider.BoundingBox)
	voxelDimension := 1.0
	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)
	var runs [3]int = [3]int{int(delta[0] / voxelDimension), int(delta[1] / voxelDimension), int(delta[2] / voxelDimension)}

	inputWorkCount := runs[0] * runs[1] * runs[2]
	inputWork := make(chan [3]int, inputWorkCount)
	workerCount := 12

	doneWorkerCount := 0
	var doneWorkerMutex sync.Mutex

	for wi := 0; wi < workerCount; wi++ {
		go func() {
			for input := range inputWork {
				i := input[0]
				j := input[1]
				k := input[2]

				voxel := &collider.BoundingBox{
					MinVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(i), float64(j), float64(k)}.Mul(voxelDimension)),
					MaxVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(i + 1), float64(j + 1), float64(k + 1)}.Mul(voxelDimension)),
				}

				// found := false
				for _, entity := range entities {
					bb := boundingBoxes[entity.GetID()]
					if collision.CheckOverlapAABBAABB(voxel, &bb) {
						outputWork <- *voxel

						// if a voxel is output, we don't need to check the rest of the entities
						// found = true
						break
					}
				}
			}

			doneWorkerMutex.Lock()
			doneWorkerCount++
			if doneWorkerCount == workerCount {
				close(outputWork)
			}
			doneWorkerMutex.Unlock()
		}()
	}

	go func() {
		for voxel := range outputWork {
			n.voxels = append(n.voxels, voxel)
			n.mutex.Lock()
			n.vertices = append(n.vertices, bbVerts(voxel)...)
			n.mutex.Unlock()
		}
		fmt.Printf("generated %d voxels\n", len(n.voxels))
	}()

	for i := 0; i < runs[0]; i++ {
		for j := 0; j < runs[1]; j++ {
			for k := 0; k < runs[2]; k++ {
				inputWork <- [3]int{i, j, k}
			}
		}
	}
	close(inputWork)

	totalTricount := 0
	for _, count := range entityTriCount {
		totalTricount += count
	}
	fmt.Println("nav mesh entity tri count", totalTricount)
}

func (n *NavigationMesh) BakeNavMesh() {
	if n.constructed {
		return
	}
	n.Voxelize()
}

func bbVerts(bb collider.BoundingBox) []mgl64.Vec3 {
	min := bb.MinVertex
	max := bb.MaxVertex
	delta := max.Sub(min)

	verts := []mgl64.Vec3{
		// top
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		max,
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),

		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),
		max,

		// bottom
		min,
		min.Add(mgl64.Vec3{delta[0], 0, 0}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),

		min,
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),
		min.Add(mgl64.Vec3{0, 0, delta[2]}),

		// left
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min,
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		min,
		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		// right
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, 0}),

		min.Add(mgl64.Vec3{delta[0], 0, 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),

		// front
		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),

		min.Add(mgl64.Vec3{0, 0, delta[2]}),
		min.Add(mgl64.Vec3{delta[0], delta[1], delta[2]}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),

		// back
		min,
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], 0, 0}),

		min,
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),
	}
	return verts
}

type AABB struct {
	Min, Max mgl64.Vec3
}

type Triangle struct {
	V1, V2, V3 mgl64.Vec3
}

func IntersectAABBTriangle(aabb AABB, tri Triangle) bool {
	// Translate AABB and triangle so that the AABB is centered at the origin.
	center := aabb.Min.Add(aabb.Max).Mul(0.5)
	extents := aabb.Max.Sub(center)

	v1 := tri.V1.Sub(center)
	v2 := tri.V2.Sub(center)
	v3 := tri.V3.Sub(center)

	// Compute edge vectors for triangle.
	e1 := v2.Sub(v1)
	e2 := v3.Sub(v2)
	e3 := v1.Sub(v3)

	// Test against axes that are parallel to the AABB's coordinate axes.
	if math.Abs(v1[0]) > extents[0] && math.Abs(v2[0]) > extents[0] && math.Abs(v3[0]) > extents[0] {
		return false
	}
	if math.Abs(v1[1]) > extents[1] && math.Abs(v2[1]) > extents[1] && math.Abs(v3[1]) > extents[1] {
		return false
	}
	if math.Abs(v1[2]) > extents[2] && math.Abs(v2[2]) > extents[2] && math.Abs(v3[2]) > extents[2] {
		return false
	}

	// Test against the plane of each face of the AABB.
	if math.Abs(v1[1]*e1[2]-v1[2]*e1[1]) > extents[1]*math.Abs(e1[2])+extents[2]*math.Abs(e1[1]) ||
		math.Abs(v2[1]*e2[2]-v2[2]*e2[1]) > extents[1]*math.Abs(e2[2])+extents[2]*math.Abs(e2[1]) ||
		math.Abs(v3[1]*e3[2]-v3[2]*e3[1]) > extents[1]*math.Abs(e3[2])+extents[2]*math.Abs(e3[1]) {
		return false
	}
	if math.Abs(v1[2]*e1[0]-v1[0]*e1[2]) > extents[0]*math.Abs(e1[2])+extents[2]*math.Abs(e1[0]) ||
		math.Abs(v2[2]*e2[0]-v2[0]*e2[2]) > extents[0]*math.Abs(e2[2])+extents[2]*math.Abs(e2[0]) ||
		math.Abs(v3[2]*e3[0]-v3[0]*e3[2]) > extents[0]*math.Abs(e3[2])+extents[2]*math.Abs(e3[0]) {
		return false
	}

	return true
}
