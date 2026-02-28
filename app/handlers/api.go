package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type JsonHandler struct {
	logger          *slog.Logger
	authService     *services.AuthService
	chatService     *services.ChatService
	profileService  *services.ProfileService
	presenceService *services.PresenceService
	pubsub          pubsub.Pubsub
}

func NewJsonHandler(
	logger *slog.Logger,
	authService *services.AuthService,
	chatService *services.ChatService,
	profileService *services.ProfileService,
	presenceService *services.PresenceService,
	pubsub pubsub.Pubsub,
) *JsonHandler {
	return &JsonHandler{
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

func (me *JsonHandler) HandleRegister(c *fiber.Ctx) error {
	var request RegisterRequest
	if err := c.BodyParser(&request); err != nil {
		return NewApiError(ErrInvalidJSON, err)
	}

	if err := me.authService.Register(request.Name, request.Username, request.Email, request.Bio); err != nil {
		return serviceErrToApiErr(err)
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

func (me *JsonHandler) HandleRequestOtp(c *fiber.Ctx) error {
	var request RequestOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return NewApiError(ErrInvalidJSON, err)
	}

	otpID, err := me.authService.SendOtp(request.Channel, request.Identifier, request.Purpose)
	if err != nil {
		return serviceErrToApiErr(err)
	}

	return c.Status(fiber.StatusOK).JSON(RequestOtpResponse{
		OtpID: otpID,
	})
}

type VerifyOtpRequest struct {
	OtpID string `json:"otpID"`
	Otp   string `json:"otp"`
}

func (me *JsonHandler) HandleVerifyOtp(c *fiber.Ctx) error {
	var request VerifyOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return NewApiError(ErrInvalidJSON, err)
	}

	platform, os := extractPlatformAndOSFromUserAgent(c.Get("User-Agent"))

	session, err := me.authService.VerifyOtp(request.OtpID, request.Otp, platform, os)
	if err != nil {
		return serviceErrToApiErr(err)
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

func (me *JsonHandler) HandleLogout(c *fiber.Ctx) error {
	if err := me.authService.DeleteSession(getCurrentUserID(c), getCurrentSessionID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return c.SendStatus(fiber.StatusOK)
}

func (me *JsonHandler) HandleDeleteSession(c *fiber.Ctx) error {
	sessionID := c.Params("session_id")
	return me.authService.DeleteSession(getCurrentUserID(c), sessionID)
}

type GetActiveSessionsResponseItem struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	Os       string `json:"os"`
}

func (me *JsonHandler) HandleGetActiveSessions(c *fiber.Ctx) error {
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

func (me *JsonHandler) HandleGetChatPartners(c *fiber.Ctx) error {
	var requestCursor GetChatsCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return NewApiError(ErrInvalidCursor, nil)
		}
	}

	limit := 15
	partners, err := me.chatService.GetChatPartners(getCurrentUserID(c), requestCursor.LastMessageIDWithLastPartner, limit)
	if err != nil {
		return serviceErrToApiErr(err)
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

func (me *JsonHandler) HandleDeleteChat(c *fiber.Ctx) error {
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

func (me *JsonHandler) HandleGetChatMessages(c *fiber.Ctx) error {
	var requestCursor GetChatMessagesCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return NewApiError(ErrInvalidCursor, nil)
		}
	}

	limit := 15
	messages, err := me.chatService.GetChatMessages(getCurrentUserID(c), c.Params("partner_id"), requestCursor.LastMessageID, limit)
	if err != nil {
		return serviceErrToApiErr(err)
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

func (me *JsonHandler) HandleMarkMessagesAsRead(c *fiber.Ctx) error {
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

func (me *JsonHandler) HandleSearchProfiles(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusOK).JSON(SearchProfileResponse{
			Items: []SearchProfileResponseItem{},
		})
	}

	var requestCursor SearchProfilesCursor
	if cq := c.Query("cursor"); cq != "" {
		if err := decodeCursor(cq, &requestCursor); err != nil {
			return NewApiError(ErrInvalidCursor, err)
		}
	}

	limit := 15
	profiles, err := me.profileService.SearchProfiles(query, limit, requestCursor.LastUserID)
	if err != nil {
		return serviceErrToApiErr(err)
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

func (me *JsonHandler) HandleGetProfile(c *fiber.Ctx) error {
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

func (me *JsonHandler) HandleGetMyProfile(c *fiber.Ctx) error {
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

func (me *JsonHandler) HandleUpdateProfile(c *fiber.Ctx) error {
	var request UpdateProfileRequest
	if err := c.BodyParser(&request); err != nil {
		return NewApiError(ErrInvalidJSON, err)
	}

	if err := me.profileService.UpdateProfile(getCurrentUserID(c), request.Name, request.Username, request.Email, request.Bio); err != nil {
		return serviceErrToApiErr(err)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (me *JsonHandler) HandleDeleteProfile(c *fiber.Ctx) error {
	if err := me.profileService.DeleteProfile(getCurrentUserID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return c.SendStatus(fiber.StatusOK)
}

// ==================== WebSocket Handler ====================

func (me *JsonHandler) HandleWebsocket(c *websocket.Conn) {
	defer c.Close()
	userID := c.Locals(currentUserID).(string)
	sessionID := c.Locals(currentSessionID).(string)

	me.logger.Info("websocket connection", "user", userID, "session", sessionID)
	defer me.logger.Info("websocket disconnection", "user", userID, "session", sessionID)

	// Defering wg.Wait() before defering cancel() is critical.
	// If not the wg.Wait() will be called before cancel(),
	// and block this go routine i.e. never close ws connection/subscribers.
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Go(func() { me.presenceService.StartHeartbeat(ctx, userID, sessionID) })

	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.PartnerPresenceEvent,
			pubsub.JsonPayloadGenerator[services.PartnerPresenceEventPayload],
			me.partnerPresenceEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.UserProfileWasUpdatedEvent,
			pubsub.JsonPayloadGenerator[services.UserProfileWasUpdatedEventPayload],
			me.userProfileWasUpdatedEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.PartnerProfileWasUpdatedEvent,
			pubsub.JsonPayloadGenerator[services.PartnerProfileWasUpdatedEventPayload],
			me.partnerProfileWasUpdatedEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.PartnerProfileWasDeletedEvent,
			pubsub.JsonPayloadGenerator[services.PartnerProfileWasDeletedEventPayload],
			me.partnerProfileWasDeletedEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.MessageWasSentEvent,
			pubsub.JsonPayloadGenerator[services.MessageWasSentEventPayload],
			me.messageWasSentEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.UserMessagesWereReadEvent,
			pubsub.JsonPayloadGenerator[services.UserMessagesWereReadEventPayload],
			me.userMessagesWereReadEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.PartnerMessagesWereReadEvent,
			pubsub.JsonPayloadGenerator[services.PartnerMessagesWereReadEventPayload],
			me.partnerMessagesWereReadEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.IncommingMessageEvent,
			pubsub.JsonPayloadGenerator[services.IncommingMessageEventPayload],
			me.incommingMessageEventHandler(userID, c),
		)
	})

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
			me.handleSendMessage(userID, message)
		default:
			me.logger.Warn("unhandeled websocket message", "kind", message.Kind, "user", userID, "session", sessionID)
		}
	}
}

func (me *JsonHandler) handleSendMessage(userID string, message WebsocketMessage) {
	if err := me.chatService.SendChatMessage(userID, message.PartnerID, message.Content, message.ClientMessageID); err != nil {
		me.logger.Error("failed to send message with chat serivce", "error", err)
	}
}

func (me *JsonHandler) partnerPresenceEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerPresenceEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      PartnerPresenceChanged,
			PartnerID: message.PartnerID,
			IsOnline:  message.IsOnline,
		})
	}
}

func (me *JsonHandler) chatWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
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

func (me *JsonHandler) userMessagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserMessagesWereReadEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:           UserMessagesWereRead,
			UserID:         message.UserID,
			ReadMessageIDs: message.ReadMessageIDs,
		})
	}
}

func (me *JsonHandler) partnerMessagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerMessagesWereReadEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:              PartnerMessagesWereRead,
			PartnerID:         message.PartnerID,
			ReadMessagesCount: message.ReadMessageCount,
		})
	}
}

func (me *JsonHandler) userProfileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserProfileWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:     UserProfileWasUpdated,
			Name:     message.Name,
			Username: message.Username,
		})
	}
}

func (me *JsonHandler) partnerProfileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerProfileWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      PartnerProfileWasUpdated,
			PartnerID: message.PartnerID,
			Name:      message.Name,
			Username:  message.Username,
		})
	}
}

func (me *JsonHandler) partnerProfileWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerProfileWasDeletedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      PartnerProfileWasDeleted,
			PartnerID: message.PartnerID,
		})
	}
}

func (me *JsonHandler) messageWasSentEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.MessageWasSentEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:            MessageWasSent,
			PartnerID:       message.PartnerID,
			MessageID:       message.MessageID,
			ClientMessageID: message.ClientMessageID,
			Content:         message.Content,
			Timestamp:       message.Timestamp,
		})
	}
}

func (me *JsonHandler) incommingMessageEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.IncommingMessageEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(WebsocketMessage{
			Kind:      IncommingMessage,
			PartnerID: message.PartnerID,
			MessageID: message.MessageID,
			Content:   message.Content,
			Timestamp: message.Timestamp,
		})
	}
}
