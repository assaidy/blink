package pubsub

// TODO: remove this. it was replaced by utils/events

import (
	"context"
	"encoding/json"
)

// MessageGenerator converts a payload into a string message to be sent over the pubsub.
type MessageGenerator func(payload any) (string, error)

// PayloadGenerator converts a received string message back into a payload.
// It is used by subscribers to deserialize the incoming message.
type PayloadGenerator func(message string) (any, error)

// PayloadHandler processes the deserialized payload.
// It is called for each message received on a subscribed event.
type PayloadHandler func(payload any) error

// Pubsub defines the interface for a publish-subscribe system.
type Pubsub interface {
	// Publish sends a message with the given event name using the provided generator.
	Publish(ctx context.Context, event string, generator MessageGenerator, payload any) error
	// Subscribe listens for messages on the specified event and blocks, processing each message with the handler.
	// The payloadGenerator is used to deserialize incoming messages before passing them to the handler.
	Subscribe(ctx context.Context, event string, payloadGenerator PayloadGenerator, handler PayloadHandler)
}

// JsonMessageGenerator serializes a payload into a JSON string message.
// It is used by publishers to convert a payload before sending over the pubsub.
func JsonMessageGenerator(payload any) (string, error) {
	bytes, err := json.Marshal(payload)
	return string(bytes), err
}

// JsonPayloadGenerator deserializes a JSON string message into the specified type T.
// It is used by subscribers to convert an incoming message back into a payload.
func JsonPayloadGenerator[T any](message string) (any, error) {
	var payload T
	err := json.Unmarshal([]byte(message), &payload)
	return payload, err
}
