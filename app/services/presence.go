package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/blink/app/utils/events"
	"github.com/valkey-io/valkey-go"
)

type PresenceService struct {
	db          *sql.DB
	queries     *repo.Queries
	logger      *slog.Logger
	cache       valkey.Client
	eventSender events.Sender
}

func NewPresenceService(db *sql.DB, cache valkey.Client, logger *slog.Logger, eventSender events.Sender) *PresenceService {
	return &PresenceService{
		db:          db,
		queries:     repo.New(db),
		cache:       cache,
		logger:      logger,
		eventSender: eventSender,
	}
}

func presenceKey(userID string) string {
	return "online:user:" + userID
}

const (
	presenceHeartbeatTick = 2 * time.Second
	offlineTimeout        = 5 * time.Second
)

var PartnerPresenceEvent = makeEventChannelForUser("PartnerPresence")

type PartnerPresenceEventPayload struct {
	PartnerID string `json:"partnerID"`
	IsOnline  bool   `json:"isOnline"`
}

func (me *PresenceService) StartHeartbeat(ctx context.Context, userID, sessionID string) {
	ticker := time.NewTicker(presenceHeartbeatTick)
	defer ticker.Stop()

	me.notifyPartnersIfPresenceChanged(context.Background(), userID, wentOnline)
	defer me.notifyPartnersIfPresenceChanged(context.Background(), userID, wentOffline)

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
			if err := me.cache.Do(context.Background(),
				me.cache.B().
					Zrem().
					Key(presenceKey(userID)).
					Member(sessionID).
					Build(),
			).Error(); err != nil {
				me.logger.Error("failed to remove session from presence", "error", err)
			}
			return
		}
	}
}

type presenceChange = bool

const (
	wentOnline  presenceChange = true
	wentOffline presenceChange = false
)

func (me *PresenceService) notifyPartnersIfPresenceChanged(ctx context.Context, userID string, change presenceChange) {
	if ok, err := me.IsUserOnline(ctx, userID); err != nil {
		me.logger.Error("failed to check if user is online", "error", err)
		return
	} else if ok {
		// User is currently online (has other active sessions).
		// No need to notify partners - they already see the user as online.
		return
	}

	partnerIDs, err := me.queries.GetAllChatPartnerIDs(ctx, userID)
	if err != nil {
		me.logger.Error("failed to get chat partner IDs", "error", err)
		return
	}

	for _, id := range partnerIDs {
		if err := events.SendJson(ctx,
			me.eventSender,
			PartnerPresenceEvent(id),
			PartnerPresenceEventPayload{
				PartnerID: userID,
				IsOnline:  change,
			}); err != nil {
			me.logger.Error("failed to send event", "error", err)
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
