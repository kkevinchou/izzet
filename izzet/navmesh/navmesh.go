package navmesh

import (
	"fmt"
	"math"
	"sync"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/collision"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/geometry"
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
	// for _, entity := range entities {
	// 	for _, rd := range entity.Model.RenderData() {
	// 		_ = rd
	// 		triCountWork <- (verts / 3)
	// 		break
	// 	}

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

				for _, entity := range entities {
					bb := boundingBoxes[entity.GetID()]
					if collision.CheckOverlapAABBAABB(voxel, &bb) {
						outputWork <- *voxel
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
}

func bbVerts(bb collider.BoundingBox) []mgl64.Vec3 {
	// top
	min := bb.MinVertex
	max := bb.MaxVertex
	delta := max.Sub(min)
	verts := []mgl64.Vec3{
		min.Add(mgl64.Vec3{0, delta[1], 0}),
		max,
		min.Add(mgl64.Vec3{delta[0], delta[1], 0}),

		min.Add(mgl64.Vec3{0, delta[1], 0}),
		min.Add(mgl64.Vec3{0, delta[1], delta[2]}),
		max,
	}
	return verts
}

type AABB struct {
	Min, Max mgl64.Vec3
}

type Triangle struct {
	V1, V2, V3 mgl64.Vec3
}

func AABBIntersectsTriangle(aabb collider.BoundingBox, tri []mgl64.Vec3) bool {
	// Calculate the center point of the AABB
	aabbCenter := aabb.MinVertex.Add(aabb.MaxVertex).Mul(0.5)

	// Calculate the half-lengths of the AABB's sides
	aabbHalfLengths := aabb.MaxVertex.Sub(aabbCenter)

	// Test the AABB against each triangle edge's axis
	edges := []mgl64.Vec3{tri[1].Sub(tri[0]), tri[2].Sub(tri[1]), tri[0].Sub(tri[2])}
	normals := []mgl64.Vec3{edges[0].Cross(edges[1]), edges[1].Cross(edges[2]), edges[2].Cross(edges[0])}

	for _, normal := range normals {
		// Project the AABB center onto the axis
		projection := aabbCenter.Dot(normal)

		// Project each triangle vertex onto the axis
		v1 := tri[0].Dot(normal)
		v2 := tri[1].Dot(normal)
		v3 := tri[2].Dot(normal)

		// Check if the AABB and the triangle overlap on this axis
		if (v1 > projection+aabbHalfLengths.Dot(normal)) || (v2 > projection+aabbHalfLengths.Dot(normal)) || (v3 > projection+aabbHalfLengths.Dot(normal)) ||
			(v1 < projection-aabbHalfLengths.Dot(normal)) || (v2 < projection-aabbHalfLengths.Dot(normal)) || (v3 < projection-aabbHalfLengths.Dot(normal)) {
			return false
		}
	}

	// The AABB and the triangle overlap on all three axes, so they intersect
	return true
}

func (n *NavigationMesh) BakeNavMesh() {
	if n.constructed {
		return
	}

	n.Voxelize()

	// spatialPartition := n.world.SpatialPartition()
	// entities := spatialPartition.QueryEntities(n.Volume)

	// n.Vertices = nil

	// var polys []*geometry.Polygon
	// tileCount := 0
	// for _, entity := range entities {
	// 	e := n.world.GetEntityByID(entity.GetID())
	// 	if !strings.Contains(e.Name, "Tile") {
	// 		continue
	// 	}

	// 	tileCount++

	// 	bb := e.BoundingBox()
	// 	if bb == nil {
	// 		continue
	// 	}

	// 	vertices := topFaceFromBoundingBox(*bb)
	// 	polys = append(polys, trianglePolygons(vertices)...)
	// 	n.Vertices = append(n.Vertices, vertices...)
	// }

	// if len(polys) > 0 {
	// 	n.navmesh, n.clusterIDs = pathing.ConstructNavMesh(polys)
	// 	n.constructed = true
	// 	clusterSet := map[int]bool{}
	// 	for _, id := range n.clusterIDs {
	// 		clusterSet[id] = true
	// 	}

	// 	fmt.Println("NUM POLYS", len(polys))
	// 	fmt.Println("NUM CLUSTERS", len(clusterSet))
	// }

	// fmt.Println(n.navmesh.PortalToPolygons())
	// fmt.Println("a")
}

func trianglePolygons(vertices []mgl64.Vec3) []*geometry.Polygon {
	var polys []*geometry.Polygon
	for i := 0; i < len(vertices); i += 3 {
		var points []geometry.Point
		verts := vertices[i : i+3]
		for _, v := range verts {

			v[0] = math.Floor(v[0])
			v[1] = math.Floor(v[1])
			v[2] = math.Floor(v[2])
			points = append(points, geometry.Point(v))
		}

		polys = append(polys, geometry.NewPolygon(points))

	}
	return polys
}

func topFaceFromBoundingBox(bb collider.BoundingBox) []mgl64.Vec3 {
	delta := bb.MaxVertex.Sub(bb.MinVertex)

	var yOffset float64 = 5

	verts := []mgl64.Vec3{
		bb.MaxVertex.Add(mgl64.Vec3{0, yOffset, 0}),
		bb.MinVertex.Add(mgl64.Vec3{delta[0], delta[1] + yOffset, 0}),
		bb.MinVertex.Add(mgl64.Vec3{0, delta[1] + yOffset, 0}),

		bb.MinVertex.Add(mgl64.Vec3{0, delta[1] + yOffset, 0}),
		bb.MaxVertex.Add(mgl64.Vec3{-delta[0], yOffset, 0}),
		bb.MaxVertex.Add(mgl64.Vec3{0, float64(yOffset), 0}),
	}

	return verts
}
