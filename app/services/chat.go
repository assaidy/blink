package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/blink/app/utils/events"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/oklog/ulid/v2"
)

type ChatService struct {
	db              *sql.DB
	queries         *repo.Queries
	presenceService *PresenceService
	eventSender     events.Sender
}

func NewChatService(db *sql.DB, presenceService *PresenceService, eventSender events.Sender) *ChatService {
	return &ChatService{
		db:              db,
		queries:         repo.New(db),
		presenceService: presenceService,
		eventSender:     eventSender,
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

var ChatWasDeletedEvent = makeEventChannelForUser("ChatWasDeleted")

type ChatWasDeletedEventPayload struct {
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

	if err := events.SendJson(ctx,
		me.eventSender,
		ChatWasDeletedEvent(userID),
		ChatWasDeletedEventPayload{
			PartnerID: partnerID,
		},
	); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	if err := events.SendJson(ctx,
		me.eventSender,
		ChatWasDeletedEvent(partnerID),
		ChatWasDeletedEventPayload{
			PartnerID: userID,
		},
	); err != nil {
		return fmt.Errorf("failed to event: %w", err)
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

var (
	UserMessagesWereReadEvent    = makeEventChannelForUser("UserMessagesWereRead")
	PartnerMessagesWereReadEvent = makeEventChannelForUser("PartnerMessagesWereRead")
)

type UserMessagesWereReadEventPayload struct {
	ReadMessageIDs []string `json:"readMessageIDs"`
}

type PartnerMessagesWereReadEventPayload struct {
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
		if err := events.SendJson(ctx,
			me.eventSender,
			UserMessagesWereReadEvent(partnerID),
			UserMessagesWereReadEventPayload{
				ReadMessageIDs: markedMessageIDs,
			},
		); err != nil {
			return fmt.Errorf("failed to event: %w", err)
		}

		if err := events.SendJson(ctx,
			me.eventSender,
			PartnerMessagesWereReadEvent(userID),
			PartnerMessagesWereReadEventPayload{
				PartnerID:        partnerID,
				ReadMessageCount: count,
			},
		); err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
	}

	return nil
}

var (
	UserMessageWasDeletedEvent    = makeEventChannelForUser("UserMessageWasDeleted")
	PartnerMessageWasDeletedEvent = makeEventChannelForUser("PartnerMessageWasDeleted")
)

type UserMessageWasDeletedEventPayload struct {
	MessageID string `json:"messageID"`
}

type PartnerMessageWasDeletedEventPayload struct {
	PartnerID string `json:"partnerID"`
	MessageID string `json:"messageID"`
}

func (me *ChatService) DeleteChatMessage(userID, messageID string) error {
	ctx := context.Background()
	tx, err := me.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()
	qtx := me.queries.WithTx(tx)

	result, err := qtx.GetChatMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to check chat message for user: %w", err)
	}

	if result.SenderID != userID {
		return ErrNotFound
	}

	if err := qtx.MarkChatMessageAsDeleted(ctx, messageID); err != nil {
		return fmt.Errorf("failed to get chat message partners: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	if err := events.SendJson(ctx,
		me.eventSender,
		UserMessageWasDeletedEvent(result.SenderID),
		UserMessageWasDeletedEventPayload{
			MessageID: messageID,
		},
	); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	if err := events.SendJson(ctx,
		me.eventSender,
		PartnerMessageWasDeletedEvent(result.ReceiverID),
		PartnerMessageWasDeletedEventPayload{
			PartnerID: result.SenderID,
			MessageID: messageID,
		},
	); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	return nil
}

var (
	UserMessageWasUpdatedEvent    = makeEventChannelForUser("UserMessageWasUpdated")
	PartnerMessageWasUpdatedEvent = makeEventChannelForUser("PartnerMessageWasUpdated")
)

type UserMessageWasUpdatedEventPayload struct {
	MessageID string `json:"messageID"`
	Content   string `json:"newContent"`
}

type PartnerMessageWasUpdatedEventPayload struct {
	PartnerID  string `json:"partnerID"`
	MessageID  string `json:"messageID"`
	NewContent string `json:"newContent"`
}

func (me *ChatService) UpdateChatMessage(userID, messageID, newContent string) error {
	newContent = strings.TrimSpace(newContent)
	if newContent == "" {
		return fmt.Errorf("%w: %w", ErrValidation, validation.Errors{"Content": errors.New("content cannot be empty")})
	}

	ctx := context.Background()
	tx, err := me.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()
	qtx := me.queries.WithTx(tx)

	result, err := qtx.GetChatMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get chat message partners: %w", err)
	}

	if result.SenderID != userID {
		return ErrNotFound
	}

	if err := qtx.UpdateChatMessageContent(ctx, repo.UpdateChatMessageContentParams{
		ID:      messageID,
		Content: newContent,
	}); err != nil {
		return fmt.Errorf("failed to update chat message content: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	if err := events.SendJson(ctx,
		me.eventSender,
		UserMessageWasUpdatedEvent(result.SenderID),
		UserMessageWasUpdatedEventPayload{
			MessageID: messageID,
			Content:   newContent,
		},
	); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	if err := events.SendJson(ctx,
		me.eventSender,
		PartnerMessageWasUpdatedEvent(result.ReceiverID),
		PartnerMessageWasUpdatedEventPayload{
			PartnerID:  result.SenderID,
			MessageID:  messageID,
			NewContent: newContent,
		},
	); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	return nil
}

var MessageWasSentEvent = makeEventChannelForUser("MessageWasSent")

type MessageWasSentEventPayload struct {
	PartnerID       string    `json:"partnerID"`
	MessageID       string    `json:"messageID"`
	ClientMessageID int       `json:"clientMessageID"`
	Content         string    `json:"content"`
	Timestamp       time.Time `json:"timestamp"`
}

var IncomingMessageEvent = makeEventChannelForUser("IncomingMessage")

type IncomingMessageEventPayload struct {
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
	if err := events.SendJson(ctx,
		me.eventSender,
		MessageWasSentEvent(senderID),
		MessageWasSentEventPayload{
			PartnerID:       receiverID,
			MessageID:       messageID,
			Content:         content,
			Timestamp:       timestamp,
			ClientMessageID: clientMessageID,
		},
	); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	// Notify receiver sessions
	if ok, err := me.presenceService.IsUserOnline(ctx, receiverID); err != nil {
		return fmt.Errorf("failed to check if user is online: %w", err)
	} else if ok {
		if err := events.SendJson(ctx,
			me.eventSender,
			IncomingMessageEvent(receiverID),
			IncomingMessageEventPayload{
				PartnerID: senderID,
				MessageID: messageID,
				Content:   content,
				Timestamp: timestamp,
			},
		); err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
	}

	return nil
}

type UnreadCount = repo.GetUnreadCountsForUserRow

func (me *ChatService) GetUnreadCounts(ctx context.Context, userID string) ([]UnreadCount, error) {
	unreadCounts, err := me.queries.GetUnreadCountsForUser(ctx, userID)
	return unreadCounts, err
}
