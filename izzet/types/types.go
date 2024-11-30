package types

type MeshHandle struct {
	Namespace string
	ID        string
}

type MaterialHandle struct {
	Namespace string
	ID        string
}

func (h MaterialHandle) String() string {
	return h.Namespace + "-" + h.ID
}
