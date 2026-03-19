package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type EventBus interface {
	Sender
	Receiver
}

type Sender interface {
	Send(ctx context.Context, channel string, payload []byte) error
}

type Handler func(ctx context.Context, payload []byte) error
type WaitUntilStop <-chan struct{}

type Receiver interface {
	Receive(ctx context.Context, channel string, handler Handler) WaitUntilStop
}

func ReceiveMany(wg *sync.WaitGroup, waits ...WaitUntilStop) {
	for _, wait := range waits {
		wg.Go(func() { <-wait })
	}
}

func SendJson(ctx context.Context, sender Sender, channel string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	return sender.Send(ctx, channel, raw)
}

type JsonHandler[T any] func(ctx context.Context, payload T) error

func ReceiveJson[T any](ctx context.Context, receiver Receiver, channel string, handler JsonHandler[T]) WaitUntilStop {
	return receiver.Receive(ctx, channel, func(ctx context.Context, payload []byte) error {
		var p T
		if err := json.Unmarshal(payload, &p); err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}
		return handler(ctx, p)
	})
}
