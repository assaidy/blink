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

type jsonHandler struct {
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
) *jsonHandler {
	return &jsonHandler{
		logger:          logger,
		authService:     authService,
		chatService:     chatService,
		profileService:  profileService,
		presenceService: presenceService,
		pubsub:          pubsub,
	}
}

// ==================== Auth API Handlers ====================

type registerRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

func (me *jsonHandler) HandleRegister(c *fiber.Ctx) error {
	var request registerRequest
	if err := c.BodyParser(&request); err != nil {
		return newApiError(errInvalidJSON, err)
	}

	if err := me.authService.Register(request.Name, request.Username, request.Email, request.Bio); err != nil {
		return serviceErrToApiErr(err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

type requestOtpRequest struct {
	Channel    string `json:"channel"`
	Identifier string `json:"identifier"`
	Purpose    string `json:"purpose"`
}

type requestOtpResponse struct {
	OtpID string `json:"otpID"`
}

func (me *jsonHandler) HandleRequestOtp(c *fiber.Ctx) error {
	var request requestOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return newApiError(errInvalidJSON, err)
	}

	otpID, err := me.authService.SendOtp(request.Channel, request.Identifier, request.Purpose)
	if err != nil {
		return serviceErrToApiErr(err)
	}

	return c.Status(fiber.StatusOK).JSON(requestOtpResponse{
		OtpID: otpID,
	})
}

type verifyOtpRequest struct {
	OtpID string `json:"otpID"`
	Otp   string `json:"otp"`
}

func (me *jsonHandler) HandleVerifyOtp(c *fiber.Ctx) error {
	var request verifyOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return newApiError(errInvalidJSON, err)
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

func (me *jsonHandler) HandleLogout(c *fiber.Ctx) error {
	if err := me.authService.DeleteSession(getCurrentUserID(c), getCurrentSessionID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return c.SendStatus(fiber.StatusOK)
}

func (me *jsonHandler) HandleDeleteSession(c *fiber.Ctx) error {
	sessionID := c.Params("session_id")
	return me.authService.DeleteSession(getCurrentUserID(c), sessionID)
}

type getActiveSessionsResponseItem struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	Os       string `json:"os"`
}

func (me *jsonHandler) HandleGetActiveSessions(c *fiber.Ctx) error {
	sessions, err := me.authService.GetActiveSessionsForUser(getCurrentUserID(c))
	if err != nil {
		return err
	}

	response := make([]getActiveSessionsResponseItem, 0, len(sessions))
	for _, s := range sessions {
		response = append(response, getActiveSessionsResponseItem{
			ID:       s.ID,
			Platform: s.Platform,
			Os:       s.Os,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ==================== Chat API Handlers ====================

type getChatsCursor struct {
	LastMessageIDWithLastPartner string `json:"lastMessageIDWithLastPartner"`
}

type getChatPartnersResponseItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	IsOnline bool   `json:"isOnline"`
}

type GetChatPartnersResponse cursoredResponse[getChatPartnersResponseItem]

func (me *jsonHandler) HandleGetChatPartners(c *fiber.Ctx) error {
	var requestCursor getChatsCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return newApiError(errInvalidCursor, err)
		}
	}

	limit := 15
	partners, err := me.chatService.GetChatPartners(getCurrentUserID(c), requestCursor.LastMessageIDWithLastPartner, limit)
	if err != nil {
		return serviceErrToApiErr(err)
	}

	var encodedResponseCursor string
	if limit == len(partners) {
		if encodedResponseCursor, err = encodeCursor(getChatsCursor{
			LastMessageIDWithLastPartner: partners[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]getChatPartnersResponseItem, 0, len(partners))
	for _, p := range partners {
		responseItems = append(responseItems, getChatPartnersResponseItem{
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

func (me *jsonHandler) HandleDeleteChat(c *fiber.Ctx) error {
	if err := me.chatService.DeleteChat(getCurrentUserID(c), c.Params("partner_id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

type getChatMessagesCursor struct {
	LastMessageID string `json:"lastMessageID"`
}

type getChatMessagesResponseItem struct {
	ID      string    `json:"id"`
	Content string    `json:"content"`
	SentAt  time.Time `json:"sentAt"`
	IsRead  bool      `json:"isRead"`
	FromMe  bool      `json:"fromMe"`
}

type GetChatMessagesResponse cursoredResponse[getChatMessagesResponseItem]

func (me *jsonHandler) HandleGetChatMessages(c *fiber.Ctx) error {
	var requestCursor getChatMessagesCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return newApiError(errInvalidCursor, err)
		}
	}

	limit := 15
	messages, err := me.chatService.GetChatMessages(getCurrentUserID(c), c.Params("partner_id"), requestCursor.LastMessageID, limit)
	if err != nil {
		return serviceErrToApiErr(err)
	}

	var encodedResponseCursor string
	if limit == len(messages) {
		if encodedResponseCursor, err = encodeCursor(getChatMessagesCursor{
			LastMessageID: messages[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]getChatMessagesResponseItem, 0, len(messages))
	for _, m := range messages {
		responseItems = append(responseItems, getChatMessagesResponseItem{
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

func (me *jsonHandler) HandleMarkMessagesAsRead(c *fiber.Ctx) error {
	if err := me.chatService.MarkMessagesAsRead(getCurrentUserID(c), c.Params("partner_id"), c.Query("upto_message_id")); err != nil {
		return serviceErrToApiErr(err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (me *jsonHandler) HandleDeleteChatMessage(c *fiber.Ctx) error {
	if err := me.chatService.DeleteChatMessage(getCurrentUserID(c), c.Params("message_id")); err != nil {
		return serviceErrToApiErr(err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (me *jsonHandler) HandleUpdateChatMessage(c *fiber.Ctx) error {
	if err := me.chatService.UpdateChatMessage(getCurrentUserID(c), c.Params("message_id"), c.FormValue("content")); err != nil {
		return serviceErrToApiErr(err)
	}
	return c.SendStatus(fiber.StatusOK)
}

// ==================== Profile API Handlers ====================

type searchProfilesCursor struct {
	LastUserID string `json:"lastUserID"`
}

type searchProfileResponseItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type SearchProfileResponse cursoredResponse[searchProfileResponseItem]

func (me *jsonHandler) HandleSearchProfiles(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusOK).JSON(SearchProfileResponse{
			Items: []searchProfileResponseItem{},
		})
	}

	var requestCursor searchProfilesCursor
	if cq := c.Query("cursor"); cq != "" {
		if err := decodeCursor(cq, &requestCursor); err != nil {
			return newApiError(errInvalidCursor, err)
		}
	}

	limit := 15
	profiles, err := me.profileService.SearchProfiles(query, limit, requestCursor.LastUserID)
	if err != nil {
		return serviceErrToApiErr(err)
	}

	var encodedResponseCursor string
	if limit == len(profiles) {
		if encodedResponseCursor, err = encodeCursor(searchProfilesCursor{
			LastUserID: profiles[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]searchProfileResponseItem, 0, len(profiles))
	for _, p := range profiles {
		responseItems = append(responseItems, searchProfileResponseItem{
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

type getProfileResponse struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Bio      string    `json:"bio"`
	JoinedAt time.Time `json:"joinedAt"`
}

func (me *jsonHandler) HandleGetProfile(c *fiber.Ctx) error {
	profile, err := me.profileService.GetProfile(c.Params("user_id"))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(getProfileResponse{
		ID:       profile.ID,
		Name:     profile.Name,
		Username: profile.Username,
		Bio:      profile.Bio,
		JoinedAt: profile.JoinedAt,
	})
}

type getMyProfileResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	EmailIsVerified bool      `json:"emailIsVerified"`
	Bio             string    `json:"bio"`
	JoinedAt        time.Time `json:"joinedAt"`
}

func (me *jsonHandler) HandleGetMyProfile(c *fiber.Ctx) error {
	profile, err := me.profileService.GetProfile(getCurrentUserID(c))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(getMyProfileResponse{
		ID:              profile.ID,
		Name:            profile.Name,
		Username:        profile.Username,
		Email:           profile.Email,
		EmailIsVerified: profile.EmailIsVerified,
		Bio:             profile.Bio,
		JoinedAt:        profile.JoinedAt,
	})
}

type updateProfileRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

func (me *jsonHandler) HandleUpdateProfile(c *fiber.Ctx) error {
	var request updateProfileRequest
	if err := c.BodyParser(&request); err != nil {
		return newApiError(errInvalidJSON, err)
	}

	if err := me.profileService.UpdateProfile(getCurrentUserID(c), request.Name, request.Username, request.Email, request.Bio); err != nil {
		return serviceErrToApiErr(err)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (me *jsonHandler) HandleDeleteProfile(c *fiber.Ctx) error {
	if err := me.profileService.DeleteProfile(getCurrentUserID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return c.SendStatus(fiber.StatusOK)
}

// ==================== WebSocket Handler ====================

func (me *jsonHandler) HandleWebsocket(c *websocket.Conn) {
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
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.UserMessageWasDeletedEvent,
			pubsub.JsonPayloadGenerator[services.UserMessageWasDeletedEventPayload],
			me.userMessageWasDeletedEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.PartnerMessageWasDeletedEvent,
			pubsub.JsonPayloadGenerator[services.PartnerMessageWasDeletedEventPayload],
			me.partnerMessageWasDeletedEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.UserMessageWasUpdatedEvent,
			pubsub.JsonPayloadGenerator[services.UserMessageWasUpdatedEventPayload],
			me.userMessageWasUpdatedEventHandler(userID, c),
		)
	})
	wg.Go(func() {
		me.pubsub.Subscribe(ctx,
			services.PartnerMessageWasUpdatedEvent,
			pubsub.JsonPayloadGenerator[services.PartnerMessageWasUpdatedEventPayload],
			me.partnerMessageWasUpdatedEventHandler(userID, c),
		)
	})

	for {
		message := websocketMessage{}
		if err := c.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				break
			}
			me.logger.Error("failed to read json from ws", "user", userID, "session", sessionID)
			continue
		}

		switch message.Kind {
		case sendMessage:
			me.handleSendMessage(userID, message)
		default:
			me.logger.Warn("unhandeled websocket message", "kind", message.Kind, "user", userID, "session", sessionID)
		}
	}
}

func (me *jsonHandler) handleSendMessage(userID string, message websocketMessage) {
	if err := me.chatService.SendChatMessage(userID, message.PartnerID, message.Content, message.ClientMessageID); err != nil {
		me.logger.Error("failed to send message with chat serivce", "error", err)
	}
}

func (me *jsonHandler) partnerPresenceEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerPresenceEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      partnerPresenceChanged,
			PartnerID: message.PartnerID,
			IsOnline:  message.IsOnline,
		})
	}
}

func (me *jsonHandler) chatWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.ChatWasDeletedEventPayload)
		if message.UserID != userID && message.PartnerID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      chatWasDeleted,
			UserID:    message.UserID,
			PartnerID: message.PartnerID,
		})
	}
}

func (me *jsonHandler) userMessagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserMessagesWereReadEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:           userMessagesWereRead,
			UserID:         message.UserID,
			ReadMessageIDs: message.ReadMessageIDs,
		})
	}
}

func (me *jsonHandler) partnerMessagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerMessagesWereReadEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:              partnerMessagesWereRead,
			PartnerID:         message.PartnerID,
			ReadMessagesCount: message.ReadMessageCount,
		})
	}
}

func (me *jsonHandler) userProfileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserProfileWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:     userProfileWasUpdated,
			Name:     message.Name,
			Username: message.Username,
		})
	}
}

func (me *jsonHandler) partnerProfileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerProfileWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      partnerProfileWasUpdated,
			PartnerID: message.PartnerID,
			Name:      message.Name,
			Username:  message.Username,
		})
	}
}

func (me *jsonHandler) partnerProfileWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerProfileWasDeletedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      partnerProfileWasDeleted,
			PartnerID: message.PartnerID,
		})
	}
}

func (me *jsonHandler) messageWasSentEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.MessageWasSentEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:            messageWasSent,
			PartnerID:       message.PartnerID,
			MessageID:       message.MessageID,
			ClientMessageID: message.ClientMessageID,
			Content:         message.Content,
			Timestamp:       message.Timestamp,
		})
	}
}

func (me *jsonHandler) incommingMessageEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.IncommingMessageEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      incommingMessage,
			PartnerID: message.PartnerID,
			MessageID: message.MessageID,
			Content:   message.Content,
			Timestamp: message.Timestamp,
		})
	}
}

func (me *jsonHandler) userMessageWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserMessageWasDeletedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      userMessageWasDeleted,
			UserID:    message.UserID,
			MessageID: message.MessageID,
		})
	}
}

func (me *jsonHandler) partnerMessageWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerMessageWasDeletedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      partnerMessageWasDeleted,
			UserID:    userID,
			PartnerID: message.PartnerID,
			MessageID: message.MessageID,
		})
	}
}
func (me *jsonHandler) userMessageWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserMessageWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      userMessageWasUpdated,
			UserID:    userID,
			MessageID: message.MessageID,
			Content:   message.Content,
		})
	}
}

func (me *jsonHandler) partnerMessageWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerMessageWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}
		return c.WriteJSON(websocketMessage{
			Kind:      partnerMessageWasUpdated,
			UserID:    userID,
			PartnerID: message.PartnerID,
			MessageID: message.MessageID,
		})
	}
}
