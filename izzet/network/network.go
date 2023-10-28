package network

import "encoding/json"

func ExtractMessage[T any](messageTransport MessageTransport) (T, error) {
	var message T
	err := json.Unmarshal(messageTransport.Body, &message)
	if err != nil {
		var empty T
		return empty, err
	}

	return message, nil
}
