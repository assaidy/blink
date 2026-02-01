package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type WebsocketHandler struct {
	logger          *slog.Logger
	chatService     *services.ChatService
	presenceService *services.PresenceService
	pubsub          pubsub.Pubsub
}

func NewWebsocketHandler(logger *slog.Logger, presenceService *services.PresenceService) *WebsocketHandler {
	return &WebsocketHandler{logger: logger}
}

func (me *WebsocketHandler) WithWebsocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return utils.NewError(utils.WebscoketUpgradeRequired, nil)
}

const (
	ChatPartnerPresenceChanged = "ChatPartnerPresenceChanged"
	ChatWasDeleted             = "ChatWasDeleted"
	MessagesWereRead           = "MessagesWereRead"
	ProfileWasUpdated          = "ProfileWasUpdated"
	ProfileWasDeleted          = "ProfileWasDeleted"
	SendMessage                = "SendMessage"
)

type WebsocketMessage struct {
	Kind            string    `json:"kind"`
	UserID          string    `json:"userID,omitempty"`
	PartnerID       string    `json:"partnerID,omitempty"`
	IsOnline        bool      `json:"isOnline,omitempty"`
	UptoMessageID   string    `json:"uptoMessageID,omitempty"`
	Name            string    `json:"name,omitempty"`
	Email           string    `json:"email,omitempty"`
	Bio             string    `json:"bio,omitempty"`
	MessageID       string    `json:"messageID"`
	ClientMessageID int       `json:"clientMessageID"`
	Content         string    `json:"content"`
	Timestamp       time.Time `json:"timestamp"`
}

func (me *WebsocketHandler) HandleApiWebsocket(c *websocket.Conn) {
	defer c.Close()
	userID := c.Locals(currentUserID).(string)
	sessionID := c.Locals(currentSessionID).(string)

	me.logger.Info("websocket connection", "user", userID, "session", sessionID)
	defer me.logger.Info("websocket disconnection", "user", userID, "session", sessionID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go me.presenceService.StartHeartbeat(ctx, userID, sessionID)

	go me.pubsub.Subscribe(ctx,
		services.ChatPartnerPresenceEvent,
		pubsub.JsonPayloadGenerator[services.ChatPartnerPresenceEventPayload],
		me.chatPartnerPresenceEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.ChatWasDeletedEvent,
		pubsub.JsonPayloadGenerator[services.ChatWasDeletedEventPayload],
		me.chatWasDeletedEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.MessagesWereReadEvent,
		pubsub.JsonPayloadGenerator[services.MessagesWereReadEventPayload],
		me.messagesWereReadEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.ProfileWasUpdatedEvent,
		pubsub.JsonPayloadGenerator[services.ProfileWasUpdatedEventPayload],
		me.profileWasUpdatedEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.ProfileWasDeletedEvent,
		pubsub.JsonPayloadGenerator[services.ProfileWasDeletedEventPayload],
		me.profileWasDeletedEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.MessageWasSentEvent,
		pubsub.JsonPayloadGenerator[services.MessageWasSentEventPayload],
		me.MessageWasSentEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.IncommingMessageEvent,
		pubsub.JsonPayloadGenerator[services.IncommingMessageEventPayload],
		me.IncommingMessageEventHandler(userID, c),
	)

	for {
		message := WebsocketMessage{}
		if err := c.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				break
			}
			me.logger.Error("failed to read json from ws", "user", userID, "session", sessionID)
			continue
		}

		switch message.Kind {
		case SendMessage:
			if err := me.chatService.SendChatMessage(userID, message.PartnerID, message.Content, message.ClientMessageID); err != nil {
				me.logger.Error("failed to send message with chat serivce", "error", err)
			}
		default:
			me.logger.Warn("unhandeled websocket message", "kind", message.Kind, "user", userID, "session", sessionID)
		}
	}
}

func (me *WebsocketHandler) chatPartnerPresenceEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.ChatPartnerPresenceEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      ChatPartnerPresenceChanged,
			PartnerID: message.PartnerID,
			IsOnline:  message.IsOnline,
		})
	}
}

func (me *WebsocketHandler) chatWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.ChatWasDeletedEventPayload)
		if message.UserID != userID && message.PartnerID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      ChatWasDeleted,
			UserID:    message.UserID,
			PartnerID: message.PartnerID,
		})
	}
}

func (me *WebsocketHandler) messagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.MessagesWereReadEventPayload)
		if message.UserID != userID && message.PartnerID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:          MessagesWereRead,
			UserID:        message.UserID,
			PartnerID:     message.PartnerID,
			UptoMessageID: message.UptoMessageID,
		})
	}
}

func (me *WebsocketHandler) profileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.ProfileWasUpdatedEventPayload)
		if message.UserID != userID && message.PartnerID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      ProfileWasUpdated,
			UserID:    message.UserID,
			PartnerID: message.PartnerID,
			Name:      message.Name,
			Email:     message.Email,
			Bio:       message.Bio,
		})
	}
}

func (me *WebsocketHandler) profileWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.ProfileWasDeletedEventPayload)
		if message.UserID != userID && message.PartnerID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      ProfileWasDeleted,
			UserID:    message.UserID,
			PartnerID: message.PartnerID,
		})
	}
}

func (me *WebsocketHandler) MessageWasSentEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.MessageWasSentEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:            ProfileWasDeleted,
			PartnerID:       message.PartnerID,
			MessageID:       message.MessageID,
			ClientMessageID: message.ClientMessageID,
			Content:         message.Content,
			Timestamp:       message.Timestamp,
		})
	}
}

func (me *WebsocketHandler) IncommingMessageEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.IncommingMessageEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      ProfileWasDeleted,
			PartnerID: message.PartnerID,
			MessageID: message.MessageID,
			Content:   message.Content,
			Timestamp: message.Timestamp,
		})
	}
}
