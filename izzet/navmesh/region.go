package navmesh

func BuildRegions(chf *CompactHeightField) {
	for z := range chf.height {
		for x := range chf.width {
			cell := &chf.cells[x+z*chf.width]
			spanIndex := cell.SpanIndex
			spanCount := cell.SpanCount

			for i := spanIndex; i < spanIndex+spanCount; i++ {
				span := &chf.spans[i]
				for _, dir := range dirs {
					neighborX := x + xDirs[dir]
					neighborZ := z + zDirs[dir]

					if neighborX < 0 || neighborZ < 0 || neighborX >= chf.width || neighborZ >= chf.height {
						continue
					}
				}
			}
		}
	}
}
