package handle

import "encoding/json"

type Material struct {
	id string
}

func NewMaterial(id string) Material {
	return Material{id: id}
}

func (h Material) ID() string {
	return h.id
}

func (h Material) String() string {
	return h.id
}

func (h Material) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID string
	}{
		ID: h.id,
	})
}

func (h *Material) UnmarshalJSON(data []byte) error {
	var value struct {
		ID string
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = NewMaterial(value.ID)
	return nil
}
