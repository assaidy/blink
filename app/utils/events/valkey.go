package events

import (
	"context"
	"errors"
	"log/slog"

	"github.com/valkey-io/valkey-go"
)

type ValkeySenderReceiver struct {
	client valkey.Client
	logger *slog.Logger
}

func NewValkeySenderReceiver(client valkey.Client, logger *slog.Logger) ValkeySenderReceiver {
	return ValkeySenderReceiver{client: client, logger: logger}
}

func (me ValkeySenderReceiver) Send(ctx context.Context, channel string, payload []byte) error {
	cmd := me.client.B().Publish().Channel(channel).Message(string(payload)).Build()
	return me.client.Do(ctx, cmd).Error()
}

func (me ValkeySenderReceiver) Receive(ctx context.Context, channel string, handler Handler) {
	cmd := me.client.B().Subscribe().Channel(channel).Build()
	if err := me.client.Receive(ctx, cmd, func(msg valkey.PubSubMessage) {
		if err := handler(ctx, []byte(msg.Message)); err != nil {
			me.logger.Error("failed to handle payload in valkey pubsub", "error", err)
		}
	}); err != nil && !errors.Is(err, context.Canceled) {
		me.logger.Error("failed to receive message in valkey pubsub", "error", err)
	}
}
