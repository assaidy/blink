package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type ApiHandler struct {
	logger          *slog.Logger
	authService     *services.AuthService
	chatService     *services.ChatService
	profileService  *services.ProfileService
	presenceService *services.PresenceService
	pubsub          pubsub.Pubsub
}

func NewApiHandler(
	logger *slog.Logger,
	authService *services.AuthService,
	chatService *services.ChatService,
	profileService *services.ProfileService,
	presenceService *services.PresenceService,
	pubsub pubsub.Pubsub,
) *ApiHandler {
	return &ApiHandler{
		logger:          logger,
		authService:     authService,
		chatService:     chatService,
		profileService:  profileService,
		presenceService: presenceService,
		pubsub:          pubsub,
	}
}

// ==================== Auth API Handlers ====================

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

func (me *ApiHandler) HandleRegister(c *fiber.Ctx) error {
	var request RegisterRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	if err := me.authService.Register(services.RegisterParams{
		Name:     request.Name,
		Username: request.Username,
		Email:    request.Email,
		Bio:      request.Bio,
	}); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

type RequestOtpRequest struct {
	Channel    string `json:"channel"`
	Identifier string `json:"identifier"`
	Purpose    string `json:"purpose"`
}

type RequestOtpResponse struct {
	OtpID string `json:"otpID"`
}

func (me *ApiHandler) HandleRequestOtp(c *fiber.Ctx) error {
	var request RequestOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	otpID, err := me.authService.SendOtp(services.SendOtpParams{
		Channel:   request.Channel,
		Identifer: request.Identifier,
		Purpose:   request.Purpose,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(RequestOtpResponse{
		OtpID: otpID,
	})
}

type VerifyOtpRequest struct {
	OtpID string `json:"otpID"`
	Otp   string `json:"otp"`
}

func (me *ApiHandler) HandleVerifyOtp(c *fiber.Ctx) error {
	var request VerifyOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	platform, os := extractPlatformAndOSFromUserAgent(c.Get("User-Agent"))

	session, err := me.authService.VerifyOtp(services.VerifyOtpParams{
		OtpID:    request.OtpID,
		Otp:      request.Otp,
		Platform: platform,
		OS:       os,
	})
	if err != nil {
		return err
	}

	if session != nil {
		c.Cookie(&fiber.Cookie{
			Name:     "session_token",
			Value:    session.SessionToken,
			Expires:  session.ExpiresAt,
			HTTPOnly: true,
		})
		c.Cookie(&fiber.Cookie{
			Name:    "csrf_token",
			Value:   session.CsrfToken,
			Expires: session.ExpiresAt,
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (me *ApiHandler) HandleLogout(c *fiber.Ctx) error {
	if err := me.authService.Logout(getCurrentSessionID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return c.SendStatus(fiber.StatusOK)
}

type GetActiveSessionsResponseItem struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	Os       string `json:"os"`
	App      string `json:"app"`
}

func (me *ApiHandler) HandleGetActiveSessions(c *fiber.Ctx) error {
	sessions, err := me.authService.GetActiveSessionsForUser(getCurrentUserID(c))
	if err != nil {
		return err
	}

	response := make([]GetActiveSessionsResponseItem, 0, len(sessions))
	for _, s := range sessions {
		response = append(response, GetActiveSessionsResponseItem{
			ID:       s.ID,
			Platform: s.Platform,
			Os:       s.Os,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ==================== Chat API Handlers ====================

type GetChatsCursor struct {
	LastMessageIDWithLastPartner string `json:"lastMessageIDWithLastPartner"`
}

type GetChatPartnersResponseItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	IsOnline bool   `json:"isOnline"`
}

type GetChatPartnersResponse CursoredResponse[GetChatPartnersResponseItem]

func (me *ApiHandler) HandleGetChatPartners(c *fiber.Ctx) error {
	var requestCursor GetChatsCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return utils.NewError(utils.InvalidCursor, nil)
		}
	}

	limit := 15
	partners, err := me.chatService.GetChatPartners(getCurrentUserID(c), requestCursor.LastMessageIDWithLastPartner, limit)
	if err != nil {
		return err
	}

	var encodedResponseCursor string
	if limit == len(partners) {
		if encodedResponseCursor, err = encodeCursor(GetChatsCursor{
			LastMessageIDWithLastPartner: partners[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]GetChatPartnersResponseItem, 0, len(partners))
	for _, p := range partners {
		responseItems = append(responseItems, GetChatPartnersResponseItem{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
			IsOnline: p.IsOnline,
		})
	}

	return c.Status(fiber.StatusOK).JSON(GetChatPartnersResponse{
		Items:  responseItems,
		Cursor: encodedResponseCursor,
	})
}

func (me *ApiHandler) HandleDeleteChat(c *fiber.Ctx) error {
	if err := me.chatService.DeleteChat(getCurrentUserID(c), c.Params("partner_id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

type GetChatMessagesCursor struct {
	LastMessageID string `json:"lastMessageID"`
}

type GetChatMessagesResponseItem struct {
	ID      string    `json:"id"`
	Content string    `json:"content"`
	SentAt  time.Time `json:"sentAt"`
	IsRead  bool      `json:"isRead"`
	FromMe  bool      `json:"fromMe"`
}

type GetChatMessagesResponse CursoredResponse[GetChatMessagesResponseItem]

func (me *ApiHandler) HandleGetChatMessages(c *fiber.Ctx) error {
	var requestCursor GetChatMessagesCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return utils.NewError(utils.InvalidCursor, nil)
		}
	}

	limit := 15
	messages, err := me.chatService.GetChatMessages(getCurrentUserID(c), c.Params("partner_id"), requestCursor.LastMessageID, limit)
	if err != nil {
		return err
	}

	var encodedResponseCursor string
	if limit == len(messages) {
		if encodedResponseCursor, err = encodeCursor(GetChatMessagesCursor{
			LastMessageID: messages[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]GetChatMessagesResponseItem, 0, len(messages))
	for _, m := range messages {
		responseItems = append(responseItems, GetChatMessagesResponseItem{
			ID:      m.ID,
			Content: m.Content,
			SentAt:  m.SentAt,
			IsRead:  m.IsRead,
			FromMe:  m.FromMe,
		})
	}

	return c.Status(fiber.StatusOK).JSON(GetChatMessagesResponse{
		Items:  responseItems,
		Cursor: encodedResponseCursor,
	})
}

func (me *ApiHandler) HandleMarkMessagesAsRead(c *fiber.Ctx) error {
	if err := me.chatService.MarkMessagesAsRead(getCurrentUserID(c), c.Params("partner_id"), c.Query("upto_message_id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

// ==================== Profile API Handlers ====================

type SearchProfilesCursor struct {
	LastUserID string `json:"lastUserID"`
}

type SearchProfileResponseItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type SearchProfileResponse CursoredResponse[SearchProfileResponseItem]

func (me *ApiHandler) HandleSearchProfiles(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusOK).JSON(SearchProfileResponse{
			Items: []SearchProfileResponseItem{},
		})
	}

	var requestCursor SearchProfilesCursor
	if cq := c.Query("cursor"); cq != "" {
		if err := decodeCursor(cq, &requestCursor); err != nil {
			return utils.NewError(utils.InvalidCursor, err)
		}
	}

	limit := 15
	profiles, err := me.profileService.SearchProfiles(query, limit, requestCursor.LastUserID)
	if err != nil {
		return err
	}

	var encodedResponseCursor string
	if limit == len(profiles) {
		if encodedResponseCursor, err = encodeCursor(SearchProfilesCursor{
			LastUserID: profiles[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]SearchProfileResponseItem, 0, len(profiles))
	for _, p := range profiles {
		responseItems = append(responseItems, SearchProfileResponseItem{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
		})
	}

	return c.Status(fiber.StatusOK).JSON(SearchProfileResponse{
		Items:  responseItems,
		Cursor: encodedResponseCursor,
	})
}

type GetProfileResponse struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Bio      string    `json:"bio"`
	JoinedAt time.Time `json:"joinedAt"`
}

func (me *ApiHandler) HandleGetProfile(c *fiber.Ctx) error {
	profile, err := me.profileService.GetProfile(c.Params("user_id"))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(GetProfileResponse{
		ID:       profile.ID,
		Name:     profile.Name,
		Username: profile.Username,
		Bio:      profile.Bio,
		JoinedAt: profile.JoinedAt,
	})
}

type GetMyProfileResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	EmailIsVerified bool      `json:"emailIsVerified"`
	Bio             string    `json:"bio"`
	JoinedAt        time.Time `json:"joinedAt"`
}

func (me *ApiHandler) HandleGetMyProfile(c *fiber.Ctx) error {
	profile, err := me.profileService.GetProfile(getCurrentUserID(c))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(GetMyProfileResponse{
		ID:              profile.ID,
		Name:            profile.Name,
		Username:        profile.Username,
		Email:           profile.Email,
		EmailIsVerified: profile.EmailIsVerified,
		Bio:             profile.Bio,
		JoinedAt:        profile.JoinedAt,
	})
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

func (me *ApiHandler) HandleUpdateProfile(c *fiber.Ctx) error {
	var request UpdateProfileRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	if err := me.profileService.UpdateProfile(getCurrentUserID(c), services.UpdateProfileParams{
		Name:     request.Name,
		Username: request.Username,
		Email:    request.Email,
		Bio:      request.Bio,
	}); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (me *ApiHandler) HandleDeleteProfile(c *fiber.Ctx) error {
	if err := me.profileService.DeleteProfile(getCurrentUserID(c)); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

// ==================== WebSocket Handler ====================

func (me *ApiHandler) WithWebsocket(c *fiber.Ctx) error {
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

func (me *ApiHandler) HandleWebsocket(c *websocket.Conn) {
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
		me.messageWasSentEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.IncommingMessageEvent,
		pubsub.JsonPayloadGenerator[services.IncommingMessageEventPayload],
		me.incommingMessageEventHandler(userID, c),
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

func (me *ApiHandler) chatPartnerPresenceEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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

func (me *ApiHandler) chatWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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

func (me *ApiHandler) messagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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

func (me *ApiHandler) profileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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

func (me *ApiHandler) profileWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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

func (me *ApiHandler) messageWasSentEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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

func (me *ApiHandler) incommingMessageEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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
