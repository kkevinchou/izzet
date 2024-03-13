package navmesh

const maxStacks int = 8

// func BuildRegions(chf *CompactHeightField) {
// 	level := (chf.maxDistance + 1) & ^1
// 	stackID := -1

// 	for level > 0 {
// 		level -= 2
// 		stackID = (stackID + 1) % maxStacks

// 		if stackID == 0 {

// 		}
// 	}
// }

// func sortCellsByLevel(startLevel int, chf *CompactHeightField, regions []int, maxStacks int, lvlStacks []LevelStackEntry, logLevelsPerStack int) {
// 	for z := range chf.height {
// 		for x := range chf.width {
// 			cell := &chf.cells[x+z*chf.width]
// 			spanIndex := cell.SpanIndex
// 			spanCount := cell.SpanCount

// 			for i := spanIndex; i < spanIndex+spanCount; i++ {
// 				span := &chf.spans[i]

// 			}
// 		}
// 	}
// }

// type LevelStackEntry struct {
// 	x, y, index int
// }
