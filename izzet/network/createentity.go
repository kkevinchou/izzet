package network

type CreateEntityMessage struct {
	// Position    mgl64.Vec3
	// Orientation mgl64.Quat
	// Scale       mgl64.Vec3
	// EntityID    int
	OwnerID     int
	EntityBytes []byte
}
