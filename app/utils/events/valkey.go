package events

import (
	"context"
	"log/slog"

	"github.com/valkey-io/valkey-go"
)

type ValkeyEventBus struct {
	client valkey.Client
	logger *slog.Logger
}

func NewValkeyEventBus(client valkey.Client, logger *slog.Logger) ValkeyEventBus {
	return ValkeyEventBus{client: client, logger: logger}
}

func (me ValkeyEventBus) Send(ctx context.Context, channel string, payload []byte) error {
	cmd := me.client.B().Publish().Channel(channel).Message(string(payload)).Build()
	return me.client.Do(ctx, cmd).Error()
}

func (me ValkeyEventBus) Receive(ctx context.Context, channel string, handler Handler) WaitUntilStop {
	stopped := make(chan struct{}, 1)

	go func() {
		cmd := me.client.B().Subscribe().Channel(channel).Build()
		if err := me.client.Receive(ctx, cmd, func(msg valkey.PubSubMessage) {
			if err := handler(ctx, []byte(msg.Message)); err != nil {
				me.logger.Error("failed to handle payload in valkey pubsub", "error", err)
			}
		}); err != nil {
			me.logger.Error("failed to receive messages in valkey pubsub", "error", err)
		}

		stopped <- struct{}{}
	}()

	return stopped
}
