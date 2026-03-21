package pubsub

import (
	"context"
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

func (me ValkeyPubsub) Publish(ctx context.Context, channel string, payload []byte) error {
	cmd := me.client.B().Publish().Channel(channel).Message(string(payload)).Build()
	return me.client.Do(ctx, cmd).Error()
}

func (me ValkeyPubsub) Subscribe(ctx context.Context, channel string, handler Handler) WaitChannel {
	stopped := make(chan struct{}, 1)

	go func() {
		cmd := me.client.B().Subscribe().Channel(channel).Build()
		if err := me.client.Receive(ctx, cmd, func(msg valkey.PubSubMessage) {
			if err := handler(ctx, []byte(msg.Message)); err != nil {
				me.logger.Error("failed to handle payload in valkey pubsub", "channel", channel, "error", err)
			}
		}); err != nil {
			me.logger.Error("failed to receive messages in valkey pubsub", "channel", channel, "error", err)
		}

		stopped <- struct{}{}
	}()

	return stopped
}
