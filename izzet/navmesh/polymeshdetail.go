package navmesh

type HeightPatch struct {
	xmin   int
	zmin   int
	width  int
	height int

	data []int
}

type Bound struct {
	xmin int
	xmax int
	zmin int
	zmax int
}

type DetailedMesh struct {
}

func BuildPolyMeshDetail(mesh *Mesh, chf *CompactHeightField) *DetailedMesh {
	bounds := make([]Bound, len(mesh.polygons))
	var maxhw, maxhh int
	var nPolyVerts int

	// find max size for a polygon area
	for i := range mesh.polygons {
		polygon := &mesh.polygons[i]

		xmin := chf.width
		xmax := 0
		zmin := chf.height
		zmax := 0

		for _, vertIndex := range polygon.verts {
			vert := mesh.vertices[vertIndex]
			xmin = min(xmin, vert.X)
			xmax = max(xmax, vert.X)
			zmin = min(zmin, vert.Z)
			zmax = max(zmax, vert.Z)
			nPolyVerts++
		}

		xmin = max(0, xmin-1)
		xmax = min(chf.width, xmax+1)
		zmin = max(0, zmin-1)
		zmax = min(chf.height, zmax+1)

		bounds[i].xmin = xmin
		bounds[i].xmax = xmax
		bounds[i].zmin = zmin
		bounds[i].zmax = zmax

		if xmin >= xmax || zmin >= zmax {
			continue
		}

		maxhw = max(maxhw, xmax-xmin)
		maxhh = max(maxhh, zmax-zmin)
	}

	var hp HeightPatch
	hp.data = make([]int, maxhw*maxhh)

	// vcap := nPolyVerts + (nPolyVerts / 2)
	// tcap := vcap * 2

	var dmesh DetailedMesh
	var poly []PolyVertex

	for i := range mesh.polygons {
		polygon := &mesh.polygons[i]
		for _, vertIndex := range polygon.verts {
			vert := mesh.vertices[vertIndex]
			// TODO: this should be dynamic and come from another datastructure (chf, or mesh, or w/e)
			cs := 1
			ch := 1
			poly = append(poly, PolyVertex{
				X: vert.X * cs, Y: vert.Y * ch, Z: vert.Z * cs,
			})
		}

		hp.xmin = bounds[i].xmin
		hp.zmin = bounds[i].zmin
		hp.width = bounds[i].xmax - bounds[i].xmin
		hp.height = bounds[i].zmax - bounds[i].zmin
		getHeightData()
		buildPolyDetail()
	}

	return &dmesh
}

func getHeightData() {

}

func buildPolyDetail() {

}
