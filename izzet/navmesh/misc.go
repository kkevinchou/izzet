package navmesh

func (n *NavigationMesh2) VoxelDimension() float64 {
	return n.voxelDimension
}

func (n *NavigationMesh2) VoxelField() [][][]Voxel {
	return n.voxelField
}

func (n *NavigationMesh2) VoxelCount() int {
	return n.voxelCount
}

func findSeeds(voxelField [][][]Voxel, reachField [][][]ReachInfo, dimensions [3]int) {
	// printRegionMin := mgl32.Vec3{50, 25, 135}
	// printRegionMax := mgl32.Vec3{61, 25, 147}

	for y := 0; y < dimensions[1]; y++ {
		for z := 0; z < dimensions[2]; z++ {
			for x := 0; x < dimensions[0]; x++ {
				voxel := &voxelField[x][y][z]
				// if x >= int(printRegionMin.X()) && x <= int(printRegionMax.X()) {
				// 	if y >= int(printRegionMin.Y()) && y <= int(printRegionMax.Y()) {
				// 		if z >= int(printRegionMin.Z()) && z <= int(printRegionMax.Z()) {
				// 			if voxel.Filled {
				// 				fmt.Printf("[%5.3f] ", voxel.DistanceField)
				// 			} else {
				// 				fmt.Printf("[-----] ")
				// 			}
				// 			if voxel.X == int(printRegionMax.X()) {
				// 				fmt.Printf("\n")
				// 			}
				// 		}
				// 	}
				// }
				if !voxel.Filled {
					continue
				}

				isSeed := true
				neighbors := getNeighbors(x, y, z, voxelField, reachField, dimensions)
				for _, neighbor := range neighbors {
					if neighbor.DistanceField > voxel.DistanceField {
						isSeed = false
					}
				}
				if isSeed {
					voxel.Seed = true
				}
			}
		}
	}
}
