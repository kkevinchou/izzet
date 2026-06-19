package handle

type Mesh struct {
	Namespace string
	ID        string
}

type Material struct {
	ID string
}

func (h Material) String() string {
	return h.ID
}

type Animation string

func (h Animation) String() string {
	return string(h)
}
