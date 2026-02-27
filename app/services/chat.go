package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/oklog/ulid/v2"
)

type ChatService struct {
	db              *sql.DB
	queries         *repo.Queries
	presenceService *PresenceService
	pubsub          pubsub.Pubsub
}

func NewChatService(db *sql.DB, presenceService *PresenceService, pubsub pubsub.Pubsub) *ChatService {
	return &ChatService{
		db:              db,
		queries:         repo.New(db),
		presenceService: presenceService,
		pubsub:          pubsub,
	}
}

type ChatPartner struct {
	ID            string
	Name          string
	Username      string
	LastMessageID string
	IsOnline      bool
}

func (me *ChatService) GetChatPartners(userID, lastMessageIDWithLastPartner string, limit int) ([]ChatPartner, error) {
	ctx := context.Background()
	partners, err := me.queries.GetChats(ctx, repo.GetChatsParams{
		UserID:                       userID,
		LastMessageIDWithLastPartner: lastMessageIDWithLastPartner,
		Limit:                        int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get partners: %w", err)
	}

	result := make([]ChatPartner, 0, len(partners))
	for _, p := range partners {
		isOnline, err := me.presenceService.IsUserOnline(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check if user online: %w", err)
		}

		result = append(result, ChatPartner{
			ID:            p.ID,
			Name:          p.Name,
			Username:      p.Username,
			LastMessageID: p.LastMessageID,
			IsOnline:      isOnline,
		})
	}

	return result, nil
}

const ChatWasDeletedEvent = "ChatWasDeletedEvent"

type ChatWasDeletedEventPayload struct {
	UserID    string `json:"userID"`
	PartnerID string `json:"partnerID"`
}

func (me *ChatService) DeleteChat(userID, partnerID string) error {
	ctx := context.Background()

	if ok, err := me.queries.CheckChatPartnerID(ctx, repo.CheckChatPartnerIDParams{
		UserID:    userID,
		PartnerID: partnerID,
	}); err != nil {
		return fmt.Errorf("failed to check partner id: %w", err)
	} else if !ok {
		return ErrNotFound
	}

	if err := me.queries.MarkChatAsDeleted(ctx, repo.MarkChatAsDeletedParams{
		UserID:    userID,
		PartnerID: partnerID,
	}); err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	if err := me.pubsub.Publish(ctx, ChatWasDeletedEvent, pubsub.JsonMessageGenerator, ChatWasDeletedEventPayload{
		UserID:    userID,
		PartnerID: partnerID,
	}); err != nil {
		return fmt.Errorf("failed to publish event %s: %w", PartnerPresenceEvent, err)
	}

	return nil
}

type Message = repo.GetChatMessagesRow

func (me *ChatService) GetChatMessages(userID, partnerID, lastMessageID string, limit int) ([]Message, error) {
	ctx := context.Background()

	if ok, err := me.queries.CheckUserID(ctx, partnerID); err != nil {
		return nil, fmt.Errorf("failed to check user id: %w", err)
	} else if !ok {
		return nil, ErrNotFound
	}

	messages, err := me.queries.GetChatMessages(ctx, repo.GetChatMessagesParams{
		UserID:        userID,
		PartnerID:     partnerID,
		LastMessageID: lastMessageID,
		Limit:         int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	return messages, nil
}

const (
	UserMessagesWereReadEvent    = "UserMessagesWereReadEvent"
	PartnerMessagesWereReadEvent = "PartnerMessagesWereReadEvent"
)

type UserMessagesWereReadEventPayload struct {
	UserID         string   `json:"userID"`
	ReadMessageIDs []string `json:"ReadMessageIDs"`
}

type PartnerMessagesWereReadEventPayload struct {
	UserID           string `json:"userID"`
	PartnerID        string `json:"partnerID"`
	ReadMessageCount int    `json:"readMessageCount"`
}

func (me *ChatService) MarkMessagesAsRead(userID, partnerID, uptoMessageID string) error {
	ctx := context.Background()

	if ok, err := me.queries.CheckChatPartnerID(ctx, repo.CheckChatPartnerIDParams{
		UserID:    userID,
		PartnerID: partnerID,
	}); err != nil {
		return fmt.Errorf("failed to check partner id: %w", err)
	} else if !ok {
		return ErrNotFound
	}

	markedMessageIDs, err := me.queries.MarkMessagesAsRead(ctx, repo.MarkMessagesAsReadParams{
		UserID:        userID,
		PartnerID:     partnerID,
		UptoMessageID: uptoMessageID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark messages as read")
	}

	if count := len(markedMessageIDs); count > 0 {
		if err := me.pubsub.Publish(ctx, UserMessagesWereReadEvent, pubsub.JsonMessageGenerator, UserMessagesWereReadEventPayload{
			UserID:         partnerID,
			ReadMessageIDs: markedMessageIDs,
		}); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", UserMessagesWereReadEvent, err)
		}

		if err := me.pubsub.Publish(ctx, PartnerMessagesWereReadEvent, pubsub.JsonMessageGenerator, PartnerMessagesWereReadEventPayload{
			UserID:           userID,
			PartnerID:        partnerID,
			ReadMessageCount: count,
		}); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", PartnerMessagesWereReadEvent, err)
		}
	}

	return nil
}

const MessageWasSentEvent = "MessageWasSentEvent"

type MessageWasSentEventPayload struct {
	UserID          string    `json:"userID"`
	PartnerID       string    `json:"partnerID"`
	MessageID       string    `json:"messageID"`
	ClientMessageID int       `json:"clientMessageID"`
	Content         string    `json:"content"`
	Timestamp       time.Time `json:"timestamp"`
}

const IncommingMessageEvent = "IncommingMessageEvent"

type IncommingMessageEventPayload struct {
	UserID    string    `json:"userID"`
	PartnerID string    `json:"partnerID"`
	MessageID string    `json:"messageID"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func (me *ChatService) SendChatMessage(senderID, receiverID, content string, clientMessageID int) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("content connot be empty")
	}

	ctx := context.Background()
	tx, err := me.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()
	qtx := me.queries.WithTx(tx)

	if ok, err := qtx.CheckUserID(ctx, senderID); err != nil {
		return fmt.Errorf("failed to check user id: %w", err)
	} else if !ok {
		return ErrUnauthorized
	}

	messageID := ulid.Make().String()
	timestamp := time.Now()
	if err := qtx.InsertChatMessage(ctx, repo.InsertChatMessageParams{
		ID:         messageID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		SentAt:     timestamp,
	}); err != nil {
		return fmt.Errorf("failed to insert chat message: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	// Notify sender sessions
	if err := me.pubsub.Publish(ctx,
		MessageWasSentEvent,
		pubsub.JsonMessageGenerator,
		MessageWasSentEventPayload{
			UserID:          senderID,
			PartnerID:       receiverID,
			MessageID:       messageID,
			Content:         content,
			Timestamp:       timestamp,
			ClientMessageID: clientMessageID,
		},
	); err != nil {
		return fmt.Errorf("failed to publish %s event: %w", MessageWasSentEvent, err)
	}

	// Notify receiver sessions
	if ok, err := me.presenceService.IsUserOnline(ctx, receiverID); err != nil {
		return fmt.Errorf("failed to check if user is online: %w", err)
	} else if ok {
		if err := me.pubsub.Publish(ctx,
			IncommingMessageEvent,
			pubsub.JsonMessageGenerator,
			IncommingMessageEventPayload{
				UserID:    receiverID,
				PartnerID: senderID,
				MessageID: messageID,
				Content:   content,
				Timestamp: timestamp,
			},
		); err != nil {
			return fmt.Errorf("failed to publish %s event: %w", MessageWasSentEvent, err)
		}
	}

	return nil
}

type UnreadCount = repo.GetUnreadCountsForUserRow

func (me *ChatService) GetUnreadCounts(ctx context.Context, userID string) ([]UnreadCount, error) {
	unreadCounts, err := me.queries.GetUnreadCountsForUser(ctx, userID)
	return unreadCounts, err
}
