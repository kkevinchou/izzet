package navmesh

import "github.com/go-gl/mathgl/mgl64"

type CompiledNavMesh struct {
	Tiles []CTile
}

type CTile struct {
	Vertices          []mgl64.Vec3
	Polygons          []CPolygon
	CDetailedPolygon  []CDetailedPolygon
	CDetailedVertices []mgl64.Vec3
}

type CPolygon struct {
	Vertices      []int
	PolyNeighbors []int
}

type CDetailedPolygon struct {
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

	// TODO - figure out how i need to store detailed vertices

	// for _, poly := range inNavMesh.DetailedMesh.PolyVertices {
	// 	for _, v := range inNavMesh.DetailedMesh.PolyVertices[poly] {
	// 		tile.Vertices = append(tile.Vertices, mgl64.Vec3{
	// 			float64(v.X), float64(v.Y), float64(v.Z),
	// 		})
	// 	}
	// }

	return nm
}
