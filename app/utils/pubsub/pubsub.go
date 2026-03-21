package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type Pubsub interface {
	Publisher
	Subscriber
}

type Publisher interface {
	Publish(ctx context.Context, channel string, payload []byte) error
}

type Handler func(ctx context.Context, payload []byte) error
type WaitChannel <-chan struct{}

type Subscriber interface {
	Subscribe(ctx context.Context, channel string, handler Handler) WaitChannel
}

func SubscribeAll(wg *sync.WaitGroup, waits ...WaitChannel) {
	for _, wait := range waits {
		wg.Go(func() { <-wait })
	}
}

func PublishJson(ctx context.Context, pub Publisher, channel string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	return pub.Publish(ctx, channel, raw)
}

type JsonHandler[T any] func(ctx context.Context, payload T) error

func SubscribeJson[T any](ctx context.Context, sub Subscriber, channel string, handler JsonHandler[T]) WaitChannel {
	return sub.Subscribe(ctx, channel, func(ctx context.Context, payload []byte) error {
		var p T
		if err := json.Unmarshal(payload, &p); err != nil {
			return fmt.Errorf("failed to unmarshal json: %w", err)
		}
		return handler(ctx, p)
	})
}
