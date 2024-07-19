package navmesh

func FilterLowHeightSpans(walkableHeight int, hf *HeightField) {
	xSize := hf.Width
	zSize := hf.Height
	for z := range zSize {
		for x := range xSize {
			for span := hf.Spans[x+z*xSize]; span != nil; span = span.Next {
				floor := span.Max
				ceiling := maxHeight
				if span.Next != nil {
					ceiling = span.Next.Min
				}
				if ceiling-floor < walkableHeight {
					span.Area = NULL_AREA
				}
			}
		}
	}
}

// func FilterLedgeSpans(walkableHeight, walkableClimb int, hf *HeightField) {

// }
