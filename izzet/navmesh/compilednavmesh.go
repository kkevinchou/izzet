package navmesh

import "github.com/go-gl/mathgl/mgl64"

type CompiledNavMesh struct {
	Tiles []CTile
}

type CTile struct {
	Vertices          []mgl64.Vec3
	Polygons          []CPolygon
	CDetailedPolygon  []CDetailedPolygon
	CDetailedVertices [][]mgl64.Vec3
}

type CPolygon struct {
	Vertices      []int
	PolyNeighbors []int
}

type CDetailedPolygon struct {
	Triangles [][3]int
}

func CompileNavMesh(inNavMesh *NavigationMesh) *CompiledNavMesh {
	nm := &CompiledNavMesh{}
	nm.Tiles = append(nm.Tiles, CTile{})
	tile := &nm.Tiles[0]

	cs := inNavMesh.Mesh.CellSize
	ch := inNavMesh.Mesh.CellHeight
	min := inNavMesh.Volume.MinVertex

	for _, v := range inNavMesh.Mesh.Vertices {
		tile.Vertices = append(tile.Vertices, mgl64.Vec3{
			min.X() + float64(v.X)*cs, min.Y() + float64(v.Y)*ch, min.Z() + float64(v.Z)*cs,
		})
	}

	for _, p := range inNavMesh.Mesh.Polygons {
		tile.Polygons = append(tile.Polygons, CPolygon{
			Vertices:      p.Verts[:],
			PolyNeighbors: p.polyNeighbor[:],
		})
	}

	tile.CDetailedVertices = make([][]mgl64.Vec3, len(inNavMesh.DetailedMesh.PolyVertices))
	for i, verts := range inNavMesh.DetailedMesh.PolyVertices {
		for _, v := range verts {
			tile.CDetailedVertices[i] = append(tile.CDetailedVertices[i], mgl64.Vec3{v.X, v.Y, v.Z})
		}
	}

	tile.CDetailedPolygon = make([]CDetailedPolygon, len(inNavMesh.DetailedMesh.PolyTriangles))
	for i, tris := range inNavMesh.DetailedMesh.PolyTriangles {
		for _, tri := range tris {
			tile.CDetailedPolygon[i].Triangles = append(tile.CDetailedPolygon[i].Triangles, [3]int{tri.A, tri.B, tri.C})
		}
	}

	return nm
}
