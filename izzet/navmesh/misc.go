package navmesh

func (n *NavigationMesh) VoxelDimension() float64 {
	return n.voxelDimension
}

func (n *NavigationMesh) VoxelField() [][][]Voxel {
	return n.voxelField
}

func (n *NavigationMesh) VoxelCount() int {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.voxelCount
}
