package knetwork

import (
	"encoding/json"

	"github.com/kkevinchou/izzet/izzet/events"
)

func Serialize(e events.Event) ([]byte, error) {
	return json.Marshal(e)
}

func Deserialize(bytes []byte, event events.Event) {
	err := json.Unmarshal(bytes, &event)
	if err != nil {
		panic(err)
	}
}
