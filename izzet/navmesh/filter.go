package navmesh

func FilterLowHangingWalkableObstacles(walkableClimb int, hf *HeightField) {
	xSize := hf.width
	zSize := hf.height

	for z := range zSize {
		for x := range xSize {
			// var previousSpan *Span
			// var previousWalkable bool

			for span := hf.spans[x+z*xSize]; span != nil; span = span.next {

				// previousSpan = span
			}
		}
	}
}
