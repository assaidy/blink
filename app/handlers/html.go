package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/assaidy/blink/app/web/components"
	. "github.com/assaidy/hyper"
	. "github.com/assaidy/hyper/htmx"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type htmlHandler struct {
	logger           *slog.Logger
	pubsub           pubsub.Pubsub
	authService      *services.AuthService
	chatService      *services.ChatService
	profileService   *services.ProfileService
	presenceService  *services.PresenceService
	mu               sync.RWMutex
	websocketsSatate map[*websocket.Conn]*websocketState
}

type websocketState struct {
	mu sync.Mutex
}

func NewHtmlHandler(
	logger *slog.Logger,
	pubsub pubsub.Pubsub,
	authService *services.AuthService,
	chatService *services.ChatService,
	profileService *services.ProfileService,
	presenceService *services.PresenceService,
) *htmlHandler {
	return &htmlHandler{
		logger:           logger,
		pubsub:           pubsub,
		authService:      authService,
		chatService:      chatService,
		profileService:   profileService,
		presenceService:  presenceService,
		websocketsSatate: make(map[*websocket.Conn]*websocketState),
	}
}

func (me *htmlHandler) HandleRegisterPage(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return Render(c, components.RegisterPage())
}

func (me *htmlHandler) HandleRegister(c *fiber.Ctx) error {
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
			return Render(c, components.RegisterForm(params))
		}

		if errors.Is(err, services.ErrUsernameConflict) {
			params.UsernameErr = "Username is taken"
			return Render(c, components.RegisterForm(params))
		}

		if errors.Is(err, services.ErrEmailConflict) {
			params.EmailErr = "Email is taken"
			return Render(c, components.RegisterForm(params))
		}

		return err
	}

	return redirect(c, "/login")
}

func (me *htmlHandler) HandleLoginPage(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return Render(c, components.LoginPage())
}

func (me *htmlHandler) HandleLogin(c *fiber.Ctx) error {
	email := c.FormValue("email")

	otpID, err := me.authService.SendOtp("email", email, "login")
	if err != nil {
		params := components.LoginFormParams{Email: email}

		if errors.Is(err, services.ErrValidation) {
			var validationErrs validation.Errors
			if errors.As(err, &validationErrs) {
				params.EmailErr = validationErrs["Identifier"]
			} else {
				me.logger.Warn("expected validation error", "found", err)
			}
			return Render(c, components.LoginForm(params))
		}

		if errors.Is(err, services.ErrNotFound) {
			params.EmailErr = "Invalid email address"
			return Render(c, components.LoginForm(params))
		}

		return err
	}

	return Render(c, components.OtpForm(components.OtpFormParams{OtpID: otpID}))
}

func (me *htmlHandler) HandleVerifyOtp(c *fiber.Ctx) error {
	otpID := c.FormValue("otpID")
	otp := c.FormValue("otp")

	platform, os := extractPlatformAndOSFromUserAgent(c.Get("User-Agent"))

	session, err := me.authService.VerifyOtp(otpID, otp, platform, os)
	if err != nil {
		params := components.OtpFormParams{
			OtpID: otpID,
			Otp:   otp,
		}

		if errors.Is(err, services.ErrInvalidOtp) {
			params.OtpErr = "Invalid code"
			return Render(c, components.OtpForm(params))
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

func (me *htmlHandler) HandleChatPage(c *fiber.Ctx) error {
	userID := getCurrentUserID(c)
	profile, err := me.profileService.GetProfile(userID)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return Render(c, components.ChatPage(components.ChatPageParams{
		User: components.UserBlockParams{
			Name:     profile.Name,
			Username: profile.Username,
		},
	}))
}

func (me *htmlHandler) HandleGetChatPartners(c *fiber.Ctx) error {
	cursor := c.Query("cursor")

	limit := 15
	partners, err := me.chatService.GetChatPartners(getCurrentUserID(c), cursor, limit)
	if err != nil {
		return err
	}

	partnerBlocks := make([]components.PartnerBlockParams, 0, len(partners))
	for _, p := range partners {
		partnerBlocks = append(partnerBlocks, components.PartnerBlockParams{
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

	return Render(c, components.PartnersList(components.PartnersListParams{
		Partners:                     partnerBlocks,
		LastMessageWithLastPartnerID: lastMessageID,
	}))
}

func (me *htmlHandler) HandleProfileModal(c *fiber.Ctx) error {
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

	return Render(c, components.ProfileModal(params))
}

func (me *htmlHandler) HandleUpdateProfile(c *fiber.Ctx) error {
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
			return Render(c, components.ProfileForm(params))
		}

		if errors.Is(err, services.ErrUsernameConflict) {
			params.UsernameErr = "Username is taken"
			return Render(c, components.ProfileForm(params))
		}

		if errors.Is(err, services.ErrEmailConflict) {
			params.EmailErr = "Email is taken"
			return Render(c, components.ProfileForm(params))
		}

		return err
	}

	return Render(c, components.ProfileForm(params))
}

func (me *htmlHandler) HandleDeleteProfile(c *fiber.Ctx) error {
	if err := me.profileService.DeleteProfile(getCurrentUserID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return redirect(c, "/")
}

func (me *htmlHandler) HandleLogout(c *fiber.Ctx) error {
	if err := me.authService.DeleteSession(getCurrentUserID(c), getCurrentSessionID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return redirect(c, "/login")
}

func (me *htmlHandler) HandleDeleteSession(c *fiber.Ctx) error {
	sessionID := c.Params("session_id")
	return me.authService.DeleteSession(getCurrentUserID(c), sessionID)
}

func (me *htmlHandler) HandleSearchModal(c *fiber.Ctx) error {
	return Render(c, components.SearchModal())
}

func (me *htmlHandler) HandleSearchUsers(c *fiber.Ctx) error {
	query := c.Query("query")
	cursor := c.Query("cursor")

	if query == "" {
		return Render(c, components.SearchResult(components.SearchResultParams{Query: query}))
	}

	limit := 15
	profiles, err := me.profileService.SearchProfiles(query, limit, cursor)
	if err != nil {
		return err
	}

	profileItems := make([]components.SearchResultItemParams, 0, len(profiles))
	for _, p := range profiles {
		profileItems = append(profileItems, components.SearchResultItemParams{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
		})
	}

	return Render(c, components.SearchResult(components.SearchResultParams{
		Query:   query,
		HasMore: len(profiles) == limit,
		Items:   profileItems,
	}))
}

func (me *htmlHandler) HandleSelectPartnerFromSearch(c *fiber.Ctx) error {
	partnerID := c.Params("partner_id")

	partnerProfile, err := me.profileService.GetProfile(partnerID)
	if err != nil {
		return err
	}

	isOnline, err := me.presenceService.IsUserOnline(c.Context(), partnerID)
	if err != nil {
		return err
	}

	return Render(c, Empty(
		Div(KV{AttrId: "search-modal", AttrHxSwapOob: SwapDelete}),

		components.ChatContainer(components.ChatContainerParams{
			Partner: components.PartnerBlockParams{
				ID:       partnerID,
				Name:     partnerProfile.Name,
				Username: partnerProfile.Username,
				IsOnline: isOnline,
			},
		}),
	))
}

func (me *htmlHandler) HandleSelectPartnerFromPartnersList(c *fiber.Ctx) error {
	partnerID := c.Params("partner_id")

	partnerProfile, err := me.profileService.GetProfile(partnerID)
	if err != nil {
		return err
	}

	isOnline, err := me.presenceService.IsUserOnline(c.Context(), partnerID)
	if err != nil {
		return err
	}

	return Render(c, components.ChatContainer(components.ChatContainerParams{
		Partner: components.PartnerBlockParams{
			ID:       partnerProfile.ID,
			Name:     partnerProfile.Name,
			Username: partnerProfile.Username,
			IsOnline: isOnline,
		},
	}))
}

func (me *htmlHandler) HandleChatMessages(c *fiber.Ctx) error {
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
			Status:  IfElse(m.IsRead, components.StatusRead, components.StatusSent),
			FromMe:  m.FromMe,
		})
	}

	return Render(c, components.ChatMessagesList(components.ChatMessagesListParams{
		PartnerID: partnerID,
		Messages:  messageItems,
		HasMore:   len(messages) == limit,
	}))
}

func (me *htmlHandler) HandleWebsocket(c *websocket.Conn) {
	me.mu.RLock()
	me.websocketsSatate[c] = &websocketState{}
	me.mu.RUnlock()

	defer func() {
		c.Close()

		me.mu.RLock()
		delete(me.websocketsSatate, c)
		me.mu.RUnlock()
	}()

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

	me.sendUnreadMessageCounts(userID, c)

	for {
		var message websocketMessage
		if err := c.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				break
			}
			me.logger.Error("failed to read json from ws", "user", userID, "session", sessionID)
			continue
		}

		switch message.Kind {
		case sendMessage:
			me.handleSendMessage(userID, c, message)
		default:
			me.logger.Warn("unhandeled websocket message", "kind", message.Kind, "user", userID, "session", sessionID)
		}
	}
}

func (me *htmlHandler) withWebsocketWriter(c *websocket.Conn, f func(w io.WriteCloser) error) error {
	// Per-socket mutex is required because NextWriter closes any existing writer,
	// causing a race condition when multiple goroutines (e.g., pubsub handlers) call it concurrently.
	state, ok := me.websocketsSatate[c]
	if !ok {
		return fmt.Errorf("couldn't find websocket")
	}
	state.mu.Lock()
	defer state.mu.Unlock()

	w, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return fmt.Errorf("failed to get next writer for ws conn: %w", err)
	}
	defer w.Close() // Flush

	return f(w)
}

func (me *htmlHandler) handleSendMessage(userID string, c *websocket.Conn, message websocketMessage) {
	if err := me.withWebsocketWriter(c, func(w io.WriteCloser) error {
		return Render(w,
			Div(KV{AttrId: "new-message-inserter-" + message.PartnerID, AttrHxSwapOob: SwapAfterEnd},
				components.ChatMessage(components.ChatMessageParams{
					ClientMessageID: message.ClientMessageID,
					PartnerID:       message.PartnerID,
					Content:         message.Content,
					FromMe:          true,
					Status:          components.StatusPending,
				}),
			),
		)
	}); err != nil {
		me.logger.Error("failed to send pending message component", "error", err)
	}

	if err := me.chatService.SendChatMessage(userID, message.PartnerID, message.Content, message.ClientMessageID); err != nil {
		me.logger.Error("failed to send message with chat serivce", "error", err)
	}
}

func (me *htmlHandler) sendUnreadMessageCounts(userID string, c *websocket.Conn) error {
	return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
		unreadCounts, err := me.chatService.GetUnreadCounts(context.Background(), userID)
		if err != nil {
			return fmt.Errorf("failed to get unread counts: %w", err)
		}

		if len(unreadCounts) > 0 {
			return Render(w,
				Div(KV{AttrId: "unread-manager-anchor", AttrHxSwapOob: SwapInnerHtml},
					MapSlice(unreadCounts, func(uc services.UnreadCount) HyperNode {
						return Script(RawText(
							fmt.Sprintf(`window.unreadManager.set("%s", %d);`, uc.PartnerID, uc.Count),
						))
					}),
				),
			)
		}

		return nil
	})
}

func (me *htmlHandler) messageWasSentEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.MessageWasSentEventPayload)
		if message.UserID != userID {
			return nil
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			profile, err := me.profileService.GetProfile(message.PartnerID)
			if err != nil {
				return err
			}

			isOnline, err := me.presenceService.IsUserOnline(context.Background(), message.PartnerID)
			if err != nil {
				return err
			}

			return Render(w, Empty(
				Div(KV{
					AttrId:        fmt.Sprintf("pending-chat-message-%d", message.ClientMessageID),
					AttrHxSwapOob: SwapDelete,
				}),

				newChatMessageResponse(newChatMessageResponseParams{
					PartnerID:        message.PartnerID,
					PartnerName:      profile.Name,
					PartnerUsername:  profile.Username,
					PartnerIsOnline:  isOnline,
					MessageID:        message.MessageID,
					MessageContent:   message.Content,
					MessageTimestamp: message.Timestamp,
					MessageIsFromMe:  true,
				}),
			))
		})
	}
}

func (me *htmlHandler) incommingMessageEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.IncommingMessageEventPayload)
		if message.UserID != userID {
			return nil
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			profile, err := me.profileService.GetProfile(message.PartnerID)
			if err != nil {
				return err
			}

			isOnline, err := me.presenceService.IsUserOnline(context.Background(), message.PartnerID)
			if err != nil {
				return err
			}

			return Render(w, newChatMessageResponse(newChatMessageResponseParams{
				PartnerID:        message.PartnerID,
				PartnerName:      profile.Name,
				PartnerUsername:  profile.Username,
				PartnerIsOnline:  isOnline,
				MessageID:        message.MessageID,
				MessageContent:   message.Content,
				MessageTimestamp: message.Timestamp,
			}))
		})
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

func newChatMessageResponse(params newChatMessageResponseParams) HyperNode {
	return Empty(
		Div(KV{AttrId: "new-message-inserter-" + params.PartnerID, AttrHxSwapOob: SwapAfterEnd},
			Div(KV{"data-partner-id": params.PartnerID},
				components.ChatMessage(components.ChatMessageParams{
					ID:        params.MessageID,
					Content:   params.MessageContent,
					SentAt:    params.MessageTimestamp,
					FromMe:    params.MessageIsFromMe,
					PartnerID: params.PartnerID,
					Status:    components.StatusSent,
				}),
			),
		),

		If(!params.MessageIsFromMe,
			Div(KV{AttrId: "unread-manager-anchor", AttrHxSwapOob: SwapInnerHtml},
				Script(RawText(`
            window.unreadManager.add("`+params.PartnerID+`", 1);
        `)),
			),
		),

		Div(KV{AttrId: "partner-" + params.PartnerID, AttrHxSwapOob: SwapDelete}),

		Div(KV{AttrId: "partners-list", AttrHxSwapOob: SwapAfterBegin},
			components.PartnersListItem(components.PartnerBlockParams{
				ID:       params.PartnerID,
				Name:     params.PartnerName,
				Username: params.PartnerUsername,
				IsOnline: params.PartnerIsOnline,
			}),
		),
	)
}

func (me *htmlHandler) userProfileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserProfileWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			return Render(w,
				Div(KV{AttrId: "user-block", AttrHxSwapOob: SwapOuterHtml},
					components.UserBlock(components.UserBlockParams{
						Name:     message.Name,
						Username: message.Username,
					}),
				),
			)
		})
	}
}

func (me *htmlHandler) partnerProfileWasUpdatedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerProfileWasUpdatedEventPayload)
		if message.UserID != userID {
			return nil
		}

		isOnline, err := me.presenceService.IsUserOnline(context.Background(), message.PartnerID)
		if err != nil {
			return err
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			return Render(w, Empty(
				Div(KV{AttrId: "partner-" + message.PartnerID, AttrHxSwapOob: SwapOuterHtml},
					components.PartnersListItem(components.PartnerBlockParams{
						ID:       message.PartnerID,
						Name:     message.Name,
						Username: message.Username,
						IsOnline: isOnline,
					}),
				),

				Div(KV{AttrId: "chat-container-header-" + message.PartnerID, AttrHxSwapOob: SwapOuterHtml},
					components.ChatContainerHeader(components.PartnerBlockParams{
						ID:       message.PartnerID,
						Name:     message.Name,
						Username: message.Username,
						IsOnline: isOnline,
					}),
				),
			))
		})
	}
}

func (me *htmlHandler) partnerProfileWasDeletedEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerProfileWasDeletedEventPayload)
		if message.UserID != userID {
			return nil
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			return Render(w, Empty(
				Div(KV{AttrId: "partner-" + message.PartnerID, AttrHxSwapOob: SwapDelete}),

				Div(KV{AttrId: "chat-container", AttrHxSwapOob: SwapInnerHtml},
					components.ChatContainerPlaceholder(),
				),
			))
		})
	}
}

func (me *htmlHandler) partnerPresenceEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerPresenceEventPayload)
		if message.UserID != userID {
			return nil
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			return Render(w, Empty(
				Div(KV{AttrId: "profile-block-presence-indicator-" + message.PartnerID, AttrHxSwapOob: SwapOuterHtml},
					components.PartnerBlockPresenceIndicator(message.PartnerID, message.IsOnline),
				),

				Div(KV{AttrId: "chat-container-presence-indicator-" + message.PartnerID, AttrHxSwapOob: SwapOuterHtml},
					components.ChatContainerPresenceIndicator(message.PartnerID, message.IsOnline),
				),
			))
		})
	}
}

func (me *htmlHandler) userMessagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.UserMessagesWereReadEventPayload)
		if message.UserID != userID {
			return nil
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			return Render(w,
				MapSlice(message.ReadMessageIDs, func(id string) HyperNode {
					return Div(KV{AttrId: "unread-message-indicator-" + id, AttrHxSwapOob: SwapOuterHtml},
						components.ReadMessageIndicator(),
					)
				}),
			)
		})
	}
}

func (me *htmlHandler) partnerMessagesWereReadEventHandler(userID string, c *websocket.Conn) pubsub.PayloadHandler {
	return func(payload any) error {
		message := payload.(services.PartnerMessagesWereReadEventPayload)
		if message.UserID != userID {
			return nil
		}

		return me.withWebsocketWriter(c, func(w io.WriteCloser) error {
			return Render(w,
				Div(KV{AttrId: "unread-manager-anchor", AttrHxSwapOob: SwapInnerHtml},
					Script(RawText(
						fmt.Sprintf(`window.unreadManager.sub("%s", %d);`, message.PartnerID, message.ReadMessageCount)),
					),
				),
			)
		})
	}
}
