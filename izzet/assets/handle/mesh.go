package handle

import "encoding/json"

type Mesh struct {
	namespace string
	id        string
}

func NewMesh(namespace string, id string) Mesh {
	return Mesh{namespace: namespace, id: id}
}

func (h Mesh) Namespace() string {
	return h.namespace
}

func (h Mesh) ID() string {
	return h.id
}

func (h Mesh) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Namespace string
		ID        string
	}{
		Namespace: h.namespace,
		ID:        h.id,
	})
}

func (h *Mesh) UnmarshalJSON(data []byte) error {
	var value struct {
		Namespace string
		ID        string
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = NewMesh(value.Namespace, value.ID)
	return nil
}
