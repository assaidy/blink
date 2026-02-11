package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/valkey-io/valkey-go"
)

type PresenceService struct {
	db      *sql.DB
	queries *repo.Queries
	logger  *slog.Logger
	cache   valkey.Client
	pubsub  pubsub.Pubsub
}

func NewPresenceService(db *sql.DB, cache valkey.Client, logger *slog.Logger, pubsub pubsub.Pubsub) *PresenceService {
	return &PresenceService{
		db:      db,
		queries: repo.New(db),
		cache:   cache,
		logger:  logger,
		pubsub:  pubsub,
	}
}

func presenceKey(userID string) string {
	return "online:user:" + userID
}

const (
	ChatPartnerPresenceEvent = "ChatPartnerPresence"

	presenceHeartbeatTick = 2 * time.Second
	offlineTimeout        = 5 * time.Second
)

type ChatPartnerPresenceEventPayload struct {
	UserID    string `json:"userID"`
	PartnerID string `json:"partnerID"`
	IsOnline  bool   `json:"isOnline"`
}

func (me *PresenceService) StartHeartbeat(ctx context.Context, userID, sessionID string) {
	ticker := time.NewTicker(presenceHeartbeatTick)
	defer ticker.Stop()

	me.notifyPartnersIfPresenceChanged(ctx, userID, true)
	defer me.notifyPartnersIfPresenceChanged(ctx, userID, false)

	key := presenceKey(userID)

	for {
		select {
		case <-ticker.C:
			if err := me.cache.Do(ctx,
				me.cache.B().
					Zadd().
					Key(key).
					ScoreMember().
					ScoreMember(float64(time.Now().UnixMilli()), sessionID).
					Build(),
			).Error(); err != nil {
				me.logger.Error("failed to heartbeat presence", "error", err)
			}

		case <-ctx.Done():
			<-time.After(presenceHeartbeatTick)
			return
		}
	}
}

// If online is true, the user went online; if false, the user went offline
func (me *PresenceService) notifyPartnersIfPresenceChanged(ctx context.Context, userID string, online bool) {
	if ok, err := me.IsUserOnline(ctx, userID); err != nil {
		me.logger.Error("failed to get check if user online", "error", err)
	} else if !ok {
		partnerIDs, err := me.queries.GetAllChatPartnerIDs(ctx, userID)
		if err != nil {
			me.logger.Error("failed to get chat partner IDs", "error", err)
		}

		for _, id := range partnerIDs {
			if me.pubsub.Publish(ctx,
				ChatPartnerPresenceEvent,
				pubsub.JsonMessageGenerator,
				ChatPartnerPresenceEventPayload{
					UserID:    id,
					PartnerID: userID,
					IsOnline:  online,
				}); err != nil {
				me.logger.Error("failed to publish event", "error", err)
			}
		}
	}
}

func (me *PresenceService) IsUserOnline(ctx context.Context, userID string) (bool, error) {
	key := presenceKey(userID)
	cutoff := time.Now().Add(-offlineTimeout).UnixMilli()

	if err := me.cache.Do(ctx,
		me.cache.B().
			Zremrangebyscore().
			Key(key).
			Min("-inf").
			Max(fmt.Sprint(cutoff)).
			Build(),
	).Error(); err != nil {
		return false, err
	}

	n, err := me.cache.Do(ctx,
		me.cache.B().
			Zcard().
			Key(key).
			Build(),
	).ToInt64()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
