package handlers

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/assaidy/blink/app/web/components"
	h "github.com/assaidy/hyper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// TODO: Might put response html using hyper directly in the handlers.

type HtmlHandler struct {
	logger          *slog.Logger
	pubsub          pubsub.Pubsub
	authService     *services.AuthService
	chatService     *services.ChatService
	profileService  *services.ProfileService
	presenceService *services.PresenceService
}

func NewHtmlHandler(
	logger *slog.Logger,
	pubsub pubsub.Pubsub,
	authService *services.AuthService,
	chatService *services.ChatService,
	profileService *services.ProfileService,
	presenceService *services.PresenceService,
) *HtmlHandler {
	return &HtmlHandler{
		logger:          logger,
		pubsub:          pubsub,
		authService:     authService,
		chatService:     chatService,
		profileService:  profileService,
		presenceService: presenceService,
	}
}

func (me *HtmlHandler) HandleRegisterPage(c *fiber.Ctx) error {
	return render(c, components.RegisterPage())
}

func (me *HtmlHandler) HandleRegister(c *fiber.Ctx) error {
	name := c.FormValue("name")
	username := c.FormValue("username")
	email := c.FormValue("email")
	bio := c.FormValue("bio")

	if err := me.authService.Register(name, username, email, bio); err != nil {
		params := components.RegisterFormParams{
			Name:     name,
			Username: username,
			Email:    email,
			Bio:      bio,
		}

		if errors.Is(err, services.ErrValidation) {
			var validationErrs validation.Errors
			if errors.As(err, &validationErrs) {
				params.NameErr = validationErrs["Name"]
				params.UsernameErr = validationErrs["Username"]
				params.EmailErr = validationErrs["Email"]
				params.BioErr = validationErrs["Bio"]
			} else {
				me.logger.Warn("expected validation error", "found", err)
			}
			return render(c, components.RegisterForm(params))
		}

		if errors.Is(err, services.ErrUsernameTaken) {
			params.UsernameErr = "Username is taken"
			return render(c, components.RegisterForm(params))
		}

		if errors.Is(err, services.ErrEmailTaken) {
			params.EmailErr = "Email is taken"
			return render(c, components.RegisterForm(params))
		}

		return err
	}

	return redirect(c, "/login")
}

func (me *HtmlHandler) HandleLoginPage(c *fiber.Ctx) error {
	return render(c, components.LoginPage())
}

func (me *HtmlHandler) HandleLogin(c *fiber.Ctx) error {
	email := c.FormValue("email")

	otpID, err := me.authService.SendOtp("email", email, "login")
	if err != nil {
		params := components.LoginFormParams{Email: email}

		if errors.Is(err, services.ErrValidation) {
			var validationErrs validation.Errors
			if errors.As(err, &validationErrs) {
				params.EmailErr = validationErrs["Email"]
			} else {
				me.logger.Warn("expected validation error", "found", err)
			}
			return render(c, components.LoginForm(params))
		}

		if errors.Is(err, services.ErrEmailNotFound) {
			params.EmailErr = "Invalid email address"
			return render(c, components.LoginForm(params))
		}

		return err
	}

	return render(c, components.OtpForm(components.OtpFormParams{OtpID: otpID}))
}

func (me *HtmlHandler) HandleVerifyOtp(c *fiber.Ctx) error {
	otpID := c.FormValue("otpID")
	otp := c.FormValue("otp")

	platform, os := extractPlatformAndOSFromUserAgent(c.Get("User-Agent"))

	session, err := me.authService.VerifyOtp(otpID, otp, platform, os)
	if err != nil {
		params := components.OtpFormParams{
			OtpID: otpID,
			Otp:   otp,
		}

		if errors.Is(err, services.ErrInvalidOTP) {
			params.OtpErr = "Invalid code"
			return render(c, components.OtpForm(params))
		}

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

	return redirect(c, "/")
}

func (me *HtmlHandler) HandleChatPage(c *fiber.Ctx) error {
	userID := getCurrentUserID(c)
	profile, err := me.profileService.GetProfile(userID)
	if err != nil {
		return err
	}

	return render(c, components.ChatPage(components.ChatPageParams{
		User: components.UserBlockParams{
			Name:     profile.Name,
			Username: profile.Username,
		},
	}))
}

func (me *HtmlHandler) HandleGetChatPartners(c *fiber.Ctx) error {
	cursor := c.Query("cursor")

	limit := 15
	partners, err := me.chatService.GetChatPartners(getCurrentUserID(c), cursor, limit)
	if err != nil {
		return err
	}

	partnerBlocks := make([]components.ProfileBlockParams, 0, len(partners))
	for _, p := range partners {
		partnerBlocks = append(partnerBlocks, components.ProfileBlockParams{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
			IsOnline: p.IsOnline,
		})
	}

	lastMessageID := ""
	if len(partners) == limit {
		lastMessageID = partners[len(partners)-1].LastMessageID
	}

	return render(c, components.PartnersList(components.PartnersListParams{
		Partners:                     partnerBlocks,
		LastMessageWithLastPartnerID: lastMessageID,
	}))
}

func (me *HtmlHandler) HandleProfileModal(c *fiber.Ctx) error {
	tab := c.Query("tab", "profile")
	params := components.ProfileModalParams{}

	switch tab {
	case "profile":
		profile, err := me.profileService.GetProfile(getCurrentUserID(c))
		if err != nil {
			return err
		}
		params.ActiveTab = components.TabProfile
		params.ProfileTabParams = components.ProfileTabParams{
			Name:            profile.Name,
			Username:        profile.Username,
			Email:           profile.Email,
			EmailIsVerified: profile.EmailIsVerified,
			Bio:             profile.Bio,
			JoinedAt:        profile.JoinedAt,
		}

	case "sessions":
		params.ActiveTab = components.TabSessions
		params.SessionsTabParams.CurrentSessionID = getCurrentSessionID(c)
		sessions, err := me.authService.GetActiveSessionsForUser(getCurrentUserID(c))
		if err != nil {
			return err
		}
		for _, session := range sessions {
			params.SessionsTabParams.Sessions = append(params.SessionsTabParams.Sessions, components.Session(session))
		}
	}

	return render(c, components.ProfileModal(params))
}

func (me *HtmlHandler) HandleUpdateProfile(c *fiber.Ctx) error {
	name := c.FormValue("name")
	username := c.FormValue("username")
	email := c.FormValue("email")
	bio := c.FormValue("bio")

	params := components.ProfileFormParams{
		Name:     name,
		Username: username,
		Email:    email,
		Bio:      bio,
	}

	if err := me.profileService.UpdateProfile(getCurrentUserID(c), name, username, email, bio); err != nil {
		if errors.Is(err, services.ErrValidation) {
			var validationErrs validation.Errors
			if errors.As(err, &validationErrs) {
				params.NameErr = validationErrs["Name"]
				params.UsernameErr = validationErrs["Username"]
				params.EmailErr = validationErrs["Email"]
				params.BioErr = validationErrs["Bio"]
			} else {
				me.logger.Warn("expected validation error", "found", err)
			}
			return render(c, components.ProfileForm(params))
		}

		if errors.Is(err, services.ErrUsernameTaken) {
			params.UsernameErr = "Username is taken"
			return render(c, components.ProfileForm(params))
		}

		if errors.Is(err, services.ErrEmailTaken) {
			params.EmailErr = "Email is taken"
			return render(c, components.ProfileForm(params))
		}

		return err
	}

	return render(c, components.ProfileForm(params))
}

func (me *HtmlHandler) HandleLogout(c *fiber.Ctx) error {
	if err := me.authService.DeleteSession(getCurrentUserID(c), getCurrentSessionID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return redirect(c, "/login")
}

func (me *HtmlHandler) HandleDeleteSession(c *fiber.Ctx) error {
	sessionID := c.Params("session_id")
	return me.authService.DeleteSession(getCurrentUserID(c), sessionID)
}

func (me *HtmlHandler) HandleSearchModal(c *fiber.Ctx) error {
	return render(c, components.SearchModal())
}

func (me *HtmlHandler) HandleSearchUsers(c *fiber.Ctx) error {
	query := c.Query("query")
	cursor := c.Query("cursor")

	if query == "" {
		return render(c, components.SearchResult(components.SearchResultParams{Query: query}))
	}

	limit := 15
	profiles, err := me.profileService.SearchProfiles(query, limit, cursor)
	if err != nil {
		return err
	}

	profileItems := make([]components.ProfileBlockParams, 0, len(profiles))
	for _, p := range profiles {
		isOnline, _ := me.presenceService.IsUserOnline(c.Context(), p.ID)

		profileItems = append(profileItems, components.ProfileBlockParams{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
			IsOnline: isOnline,
		})
	}

	return render(c, components.SearchResult(components.SearchResultParams{
		Query:          query,
		HasMore:        len(profiles) == limit,
		ProfileResults: profileItems,
	}))
}

func (me *HtmlHandler) HandleSelectPartnerFromSearch(c *fiber.Ctx) error {
	partnerID := c.Params("partner_id")

	partnerProfile, err := me.profileService.GetProfile(partnerID)
	if err != nil {
		return err
	}

	isOnline, err := me.presenceService.IsUserOnline(c.Context(), partnerID)
	if err != nil {
		return err
	}

	return render(c, h.Empty(
		h.Div(h.KV{"hx-swap-oob": "delete:#search-modal"}),

		components.ChatContainer(components.ChatContainerParams{
			Partner: components.ProfileBlockParams{
				ID:       partnerID,
				Name:     partnerProfile.Name,
				Username: partnerProfile.Username,
				IsOnline: isOnline,
			},
		}),
	))
}

func (me *HtmlHandler) HandleChatContainer(c *fiber.Ctx) error {
	partnerID := c.Params("partner_id")

	partnerProfile, err := me.profileService.GetProfile(partnerID)
	if err != nil {
		return err
	}

	isOnline, err := me.presenceService.IsUserOnline(c.Context(), partnerID)
	if err != nil {
		return err
	}

	return render(c, components.ChatContainer(components.ChatContainerParams{
		Partner: components.ProfileBlockParams{
			ID:       partnerProfile.ID,
			Name:     partnerProfile.Name,
			Username: partnerProfile.Username,
			IsOnline: isOnline,
		},
	}))
}

func (me *HtmlHandler) HandleChatMessages(c *fiber.Ctx) error {
	cursor := c.Query("cursor")
	partnerID := c.Params("partner_id")

	limit := 15
	messages, err := me.chatService.GetChatMessages(getCurrentUserID(c), partnerID, cursor, limit)
	if err != nil {
		return err
	}

	messageItems := make([]components.ChatMessageParams, 0, len(messages))
	for _, m := range messages {
		messageItems = append(messageItems, components.ChatMessageParams{
			ID:      m.ID,
			Content: m.Content,
			SentAt:  m.SentAt,
			IsRead:  m.IsRead,
			FromMe:  m.FromMe,
		})
	}

	return render(c, components.ChatMessagesList(components.ChatMessagesListParams{
		PartnerID: partnerID,
		Messages:  messageItems,
		HasMore:   len(messages) == limit,
	}))
}

func (me *HtmlHandler) HandleWebsocket(c *websocket.Conn) {
	defer c.Close()
	userID := c.Locals(currentUserID).(string)
	sessionID := c.Locals(currentSessionID).(string)

	me.logger.Info("websocket connection", "user", userID, "session", sessionID)
	defer me.logger.Info("websocket disconnection", "user", userID, "session", sessionID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go me.presenceService.StartHeartbeat(ctx, userID, sessionID)

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
	go me.pubsub.Subscribe(ctx,
		services.ProfileWasUpdatedEvent,
		pubsub.JsonPayloadGenerator[services.ProfileWasUpdatedEventPayload],
		me.profileWasUpdatedEventHandler(userID, c),
	)
	go me.pubsub.Subscribe(ctx,
		services.ChatPartnerPresenceEvent,
		pubsub.JsonPayloadGenerator[services.ChatPartnerPresenceEventPayload],
		me.chatPartnerPresenceEventHandler(userID, c),
	)

	type Message struct {
		Kind      string `json:"kind"`
		PartnerID string `json:"partnerID"`
		Content   string `json:"content"`
	}

	for {
		var message Message
		if err := c.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				break
			}
			me.logger.Error("failed to read json from ws", "user", userID, "session", sessionID)
			continue
		}

		switch message.Kind {
		case SendMessage:
			// TODO: I don't handle client message id for now. I will assume messages are sent/recived in order.
			if err := me.chatService.SendChatMessage(userID, message.PartnerID, message.Content, 0); err != nil {
				me.logger.Error("failed to send message with chat serivce", "error", err)
			}
		}
	}
}

func (me *HtmlHandler) messageWasSentEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.MessageWasSentEventPayload)
		if message.UserID != userID {
			return nil
		}

		w, err := c.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()

		profile, err := me.profileService.GetProfile(message.PartnerID)
		if err != nil {
			return err
		}

		isOnline, err := me.presenceService.IsUserOnline(context.Background(), message.PartnerID)
		if err != nil {
			return err
		}

		return h.Render(w, newChatMessageResponse(newChatMessageResponseParams{
			PartnerID:        message.PartnerID,
			PartnerName:      profile.Name,
			PartnerUsername:  profile.Username,
			PartnerIsOnline:  isOnline,
			MessageID:        message.MessageID,
			MessageContent:   message.Content,
			MessageTimestamp: message.Timestamp,
			MessageIsFromMe:  true,
		}))
	}
}

func (me *HtmlHandler) incommingMessageEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.IncommingMessageEventPayload)
		if message.UserID != userID {
			return nil
		}

		w, err := c.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()

		profile, err := me.profileService.GetProfile(message.PartnerID)
		if err != nil {
			return err
		}

		isOnline, err := me.presenceService.IsUserOnline(context.Background(), message.PartnerID)
		if err != nil {
			return err
		}

		return h.Render(w, newChatMessageResponse(newChatMessageResponseParams{
			PartnerID:        message.PartnerID,
			PartnerName:      profile.Name,
			PartnerUsername:  profile.Username,
			PartnerIsOnline:  isOnline,
			MessageID:        message.MessageID,
			MessageContent:   message.Content,
			MessageTimestamp: message.Timestamp,
		}))
	}
}

type newChatMessageResponseParams struct {
	PartnerID        string
	PartnerName      string
	PartnerUsername  string
	PartnerIsOnline  bool
	MessageID        string
	MessageContent   string
	MessageTimestamp time.Time
	MessageIsFromMe  bool
}

func newChatMessageResponse(params newChatMessageResponseParams) h.Node {
	return h.Empty(
		h.Div(h.KV{"hx-swap-oob": "afterend:#new-message-inserter"},
			h.Div(h.KV{"data-partner-id": params.PartnerID},
				components.ChatMessage(components.ChatMessageParams{
					ID:      params.MessageID,
					Content: params.MessageContent,
					SentAt:  params.MessageTimestamp,
					FromMe:  params.MessageIsFromMe,
				}),
			),
		),

		h.Div(h.KV{"hx-swap-oob": "delete:#partner-" + params.PartnerID}),

		h.Div(h.KV{"hx-swap-oob": "afterbegin:#partners-list"},
			components.PartnersListItem(components.ProfileBlockParams{
				ID:       params.PartnerID,
				Name:     params.PartnerName,
				Username: params.PartnerUsername,
				IsOnline: params.PartnerIsOnline,
			}),
		),
	)
}

func (me *HtmlHandler) profileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.ProfileWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}

		w, err := c.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()

		// It's a notfication to the user
		if message.PartnerID == "" {
			return h.Render(w, h.Div(h.KV{"hx-swap-oob": "outerHTML:#user-block"},
				components.UserBlock(components.UserBlockParams{
					Name:     message.Name,
					Username: message.Username,
				}),
			))
		}

		// It's a notfication to the partner
		isOnline, err := me.presenceService.IsUserOnline(context.Background(), message.PartnerID)
		if err != nil {
			return err
		}

		return h.Render(w, h.Empty(
			h.Div(h.KV{"hx-swap-oob": "outerHTML:#partner-" + message.PartnerID},
				components.PartnersListItem(components.ProfileBlockParams{
					ID:       message.PartnerID,
					Name:     message.Name,
					Username: message.Username,
					IsOnline: isOnline,
				}),
			),

			h.Div(h.KV{"hx-swap-oob": "outerHTML:#chat-container-header-" + message.PartnerID},
				components.ChatContainerHeader(components.ProfileBlockParams{
					ID:       message.PartnerID,
					Name:     message.Name,
					Username: message.Username,
					IsOnline: isOnline,
				}),
			),
		))
	}
}

func (me *HtmlHandler) chatPartnerPresenceEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.ChatPartnerPresenceEventPayload)
		if message.UserID != userID {
			return nil
		}

		w, err := c.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()

		return h.Render(w, h.Empty(
			h.Div(h.KV{"hx-swap-oob": "outerHTML:#profile-block-presence-indicator-" + message.PartnerID},
				components.ProfileBlockPresenceIndicator(message.PartnerID, message.IsOnline),
			),

			h.Div(h.KV{"hx-swap-oob": "outerHTML:#chat-container-presence-indicator-" + message.PartnerID},
				components.ChatContainerPresenceIndicator(message.PartnerID, message.IsOnline),
			),
		))
	}
}
