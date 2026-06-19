package handle

import "encoding/json"

type Animation struct {
	id string
}

func NewAnimation(id string) Animation {
	return Animation{id: id}
}

func (h Animation) ID() string {
	return h.id
}

func (h Animation) String() string {
	return h.id
}

func (h Animation) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.id)
}

func (h *Animation) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	*h = NewAnimation(id)
	return nil
}
