package message

import (
	"encoding/json"
	"fmt"

	"github.com/eclipse/paho.golang/paho"
)

type Message[T any] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

func Decode[T any](p *paho.Publish) (Message[T], error) {
	const op = "lib.mqtt.message.Decode"

	var payload Message[T]
	if err := json.Unmarshal(p.Payload, &payload); err != nil {
		return Message[T]{}, fmt.Errorf("%s: failed to unmarshal payload: %w", op, err)
	}

	return payload, nil
}
