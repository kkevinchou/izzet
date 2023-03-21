package navmesh

import (
	"fmt"
	"math"
	"strings"
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

	for _, entity := range sEntities {
		e := n.world.GetEntityByID(entity.GetID())
		if !strings.Contains(e.Name, "Tile") {
			continue
		}

		entities = append(entities, e)
		boundingBoxes[e.GetID()] = *e.BoundingBox()
	}

	voxelDimension := 1.0

	delta := n.Volume.MaxVertex.Sub(n.Volume.MinVertex)

	work := make(chan collider.BoundingBox)
	var wg sync.WaitGroup

	var runs [3]int = [3]int{int(delta[0] / voxelDimension), int(delta[1] / voxelDimension), int(delta[2] / voxelDimension)}

	for i := 0; i < runs[0]; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < runs[1]; j++ {
				for k := 0; k < runs[2]; k++ {
					voxel := &collider.BoundingBox{
						MinVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(index), float64(j), float64(k)}.Mul(voxelDimension)),
						MaxVertex: n.Volume.MinVertex.Add(mgl64.Vec3{float64(index + 1), float64(j + 1), float64(k + 1)}.Mul(voxelDimension)),
					}

					for _, entity := range entities {
						bb := boundingBoxes[entity.GetID()]
						if collision.CheckOverlapAABBAABB(voxel, &bb) {
							work <- *voxel
						}
					}
				}
			}
		}(i)
	}

	done := make(chan bool)
	go func() {
		for voxel := range work {
			n.voxels = append(n.voxels, voxel)
			n.vertices = append(n.vertices, bbVerts(voxel)...)
		}
		done <- true
	}()

	wg.Wait()
	close(work)

	<-done

	fmt.Println("voxel count:", len(n.voxels))
	fmt.Println("vertex count:", len(n.vertices))
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

	// var ht float32 = 1.0
	// var i float64 = 0
	// var b float64 = ((float64(i)/4) < 1.0)*1 + ((float64(i)/4) >= 1.0)*1
	// vertices := []float32{
	// 	// front
	// 	-ht, -ht, ht,
	// 	ht, -ht, ht,
	// 	ht, ht, ht,

	// 	ht, ht, ht,
	// 	-ht, ht, ht,
	// 	-ht, -ht, ht,

	// 	// back
	// 	ht, ht, -ht,
	// 	ht, -ht, -ht,
	// 	-ht, -ht, -ht,

	// 	-ht, -ht, -ht,
	// 	-ht, ht, -ht,
	// 	ht, ht, -ht,

	// 	// right
	// 	ht, -ht, ht,
	// 	ht, -ht, -ht,
	// 	ht, ht, -ht,

	// 	ht, ht, -ht,
	// 	ht, ht, ht,
	// 	ht, -ht, ht,

	// 	// left
	// 	-ht, ht, -ht,
	// 	-ht, -ht, -ht,
	// 	-ht, -ht, ht,

	// 	-ht, -ht, ht,
	// 	-ht, ht, ht,
	// 	-ht, ht, -ht,

	// 	// top
	// 	ht, ht, ht,
	// 	ht, ht, -ht,
	// 	-ht, ht, ht,

	// 	-ht, ht, ht,
	// 	ht, ht, -ht,
	// 	-ht, ht, -ht,

	// 	// bottom
	// 	-ht, -ht, ht,
	// 	ht, -ht, -ht,
	// 	ht, -ht, ht,

	// 	-ht, -ht, -ht,
	// 	ht, -ht, -ht,
	// 	-ht, -ht, ht,
	// }

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
