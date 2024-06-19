package navmesh

func FilterLowHeightSpans(walkableHeight int, hf *HeightField) {
	xSize := hf.width
	zSize := hf.height
	for z := range zSize {
		for x := range xSize {
			for span := hf.spans[x+z*xSize]; span != nil; span = span.next {
				floor := span.max
				ceiling := maxHeight
				if span.next != nil {
					ceiling = span.next.min
				}
				if ceiling-floor < walkableHeight {
					span.area = NULL_AREA
				}
			}
		}
	}
}
