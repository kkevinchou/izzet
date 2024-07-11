package navmesh

import (
	"math"
)

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

func BuildDetailedPolyMesh(mesh *Mesh, chf *CompactHeightField) *DetailedMesh {
	bounds := make([]Bound, len(mesh.polygons))
	var maxhw, maxhh int
	var nPolyVerts int

	// TODO: this should be dynamic and come from another datastructure (chf, or mesh, or w/e)
	cs := 1
	ch := 1

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

	var dmesh DetailedMesh
	var polyVerts []PolyVertex

	for i := range mesh.polygons {
		polygon := &mesh.polygons[i]
		for _, vertIndex := range polygon.verts {
			vert := mesh.vertices[vertIndex]
			polyVerts = append(polyVerts, PolyVertex{
				X: vert.X * cs, Y: vert.Y * ch, Z: vert.Z * cs,
			})
		}

		hp.xmin = bounds[i].xmin
		hp.zmin = bounds[i].zmin
		hp.width = bounds[i].xmax - bounds[i].xmin
		hp.height = bounds[i].zmax - bounds[i].zmin
		getHeightData(chf, hp, mesh.regionIDs[i])
		buildDetailedPoly(polyVerts, 10)
	}

	return &dmesh
}

type QueueItem struct {
	x, z      int
	spanIndex SpanIndex
}

func getHeightData(chf *CompactHeightField, hp HeightPatch, regionID int) {
	for i := range len(hp.data) {
		hp.data[i] = -1
	}

	var queue []QueueItem
	empty := true

	// copy the height from the same region, and mark region borders as seeds
	// to fill in the rest
	for hz := range hp.height {
		z := hp.zmin + hz
		for hx := range hp.height {
			x := hp.xmin + hx
			cell := chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+SpanIndex(spanCount); i++ {
				span := chf.spans[spanIndex]
				if span.regionID == regionID {
					hp.data[hx+hz*hp.width] = span.y
					empty = false
					border := false
					for _, dir := range dirs {
						neighborSpanIndex := span.neighbors[dir]
						if neighborSpanIndex == -1 {
							continue
						}

						neighborSpan := chf.spans[neighborSpanIndex]
						if neighborSpan.regionID != regionID {
							border = true
							break
						}
					}
					if border {
						queue = append(queue, QueueItem{x: x, z: z, spanIndex: spanIndex})
					}
					break
				}
			}
		}
	}

	if empty {
		panic("TODO: implement seedArrayWithPolyCenter")
	}

	// assume the seed is centered in the polygon, so a BFS to collect
	// height data will ensure we do not move onto overlapping polygons and
	// sample wrong heights
	for len(queue) > 0 {
		queueItem := queue[0]
		queue = queue[1:]

		x := queueItem.x
		z := queueItem.z
		spanIndex := queueItem.spanIndex

		span := chf.spans[spanIndex]

		for _, dir := range dirs {
			neighborSpanIndex := span.neighbors[dir]
			if neighborSpanIndex == -1 {
				continue
			}
			neighborX := x + xDirs[dir]
			neighborZ := z + zDirs[dir]
			hx := neighborX - hp.xmin
			hy := neighborZ - hp.zmin

			if hx >= hp.width || hy >= hp.height || hx < 0 || hy < 0 {
				continue
			}

			if hp.data[hx+hy*hp.width] != -1 {
				continue
			}

			neighborSpan := chf.spans[neighborSpanIndex]

			hp.data[hx+hy*hp.width] = neighborSpan.y
			queue = append(queue, QueueItem{x: neighborX, z: neighborZ, spanIndex: neighborSpanIndex})
		}
	}
}

func buildDetailedPoly(verts []PolyVertex, sampleDist float64) {

	// var cs float64 = 1
	// ics := 1 / cs
	minPolyExtent(verts)

	// tesselate outlines
	// this is done in a separate pass to ensure seamless height values across the poly boundaries
}

func minPolyExtent(verts []PolyVertex) float64 {
	var minDist float64 = math.MaxFloat64
	for i := range len(verts) {
		ni := (i + 1) % len(verts)
		p1 := verts[i]
		p2 := verts[ni]
		var maxEdgeDist float64 = 0
		for j := 0; j < len(verts); j++ {
			if j == i || j == ni {
				continue
			}
			d := distancePtSeg(verts[i].X, verts[i].Z, p1.X, p1.Z, p2.X, p2.Z)
			maxEdgeDist = max(maxEdgeDist, d)
		}
		minDist = min(minDist, maxEdgeDist)
	}
	return math.Sqrt(minDist)
}
