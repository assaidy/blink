package pubsub

import (
	"context"
	"encoding/json"
)

type MessageGenerator func(payload any) (string, error)
type PayloadGenerator func(message string) (any, error)
type PayloadHandler func(payload any) error

type Pubsub interface {
	Publish(ctx context.Context, event string, generator MessageGenerator, payload any) error
	// blocking
	Subscribe(ctx context.Context, event string, payloadGenerator PayloadGenerator, handler PayloadHandler)
}

func JsonMessageGenerator(payload any) (string, error) {
	bytes, err := json.Marshal(payload)
	return string(bytes), err
}

func JsonPayloadGenerator[T any](message string) (any, error) {
	var payload T
	err := json.Unmarshal([]byte(message), &payload)
	return payload, err
}
