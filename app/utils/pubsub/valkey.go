package pubsub

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/valkey-io/valkey-go"
)

type ValkeyPubsub struct {
	client valkey.Client
	logger *slog.Logger
}

func NewValkeyPubsub(client valkey.Client, logger *slog.Logger) ValkeyPubsub {
	return ValkeyPubsub{client: client, logger: logger}
}

func (me ValkeyPubsub) Publish(ctx context.Context, event string, generator MessageGenerator, payload any) error {
	message, err := generator(payload)
	if err != nil {
		return fmt.Errorf("failed to generate message from payload: %w", err)
	}
	cmd := me.client.B().Publish().Channel(event).Message(message).Build()
	return me.client.Do(ctx, cmd).Error()
}

func (me ValkeyPubsub) Subscribe(ctx context.Context, event string, generator PayloadGenerator, handler PayloadHandler) {
	cmd := me.client.B().Subscribe().Channel(event).Build()
	if err := me.client.Receive(ctx, cmd, func(msg valkey.PubSubMessage) {
		payload, err := generator(msg.Message)
		if err != nil {
			me.logger.Error("failed to generate payload in valkey pubub", "error", err)
			return
		}
		if err := handler(payload); err != nil {
			me.logger.Error("failed handle payload in valkey pubub", "error", err)
		}
	}); err != nil {
		me.logger.Error("failed to recieve message in valkey pubub", "error", err)
	}
}
