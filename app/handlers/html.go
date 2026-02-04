package handlers

import (
	"errors"
	"log/slog"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/blink/app/web/components"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

type HtmlHandler struct {
	logger         *slog.Logger
	authService    *services.AuthService
	chatService    *services.ChatService
	profileService *services.ProfileService
}

func NewHtmlHandler(
	logger *slog.Logger,
	authService *services.AuthService,
	chatService *services.ChatService,
	profileService *services.ProfileService,
) *HtmlHandler {
	return &HtmlHandler{
		logger:         logger,
		authService:    authService,
		chatService:    chatService,
		profileService: profileService,
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

	if err := me.authService.Register(services.RegisterParams{
		Name:     name,
		Username: username,
		Email:    email,
		Bio:      bio,
	}); err != nil {
		params := components.RegisterFormParams{
			Name:     name,
			Username: username,
			Email:    email,
			Bio:      bio,
		}

		var ue utils.Error
		if errors.As(err, &ue) {
			switch ue.Kind {
			case utils.InvalidData:
				if errs, ok := ue.Details.(validation.Errors); ok {
					params.NameErr = errs["Name"]
					params.UsernameErr = errs["Username"]
					params.EmailErr = errs["Email"]
					params.BioErr = errs["Bio"]
				} else {
					me.logger.Warn("expected validation error", "found", err)
				}
			case utils.UsernameConflict:
				params.UsernameErr = "Username is taken"
			case utils.EmailConflict:
				params.EmailErr = "Email is taken"
			}

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

	otpID, err := me.authService.SendOtp(services.SendOtpParams{
		Channel:   "email",
		Identifer: email,
		Purpose:   "login",
	})
	if err != nil {
		params := components.LoginFormParams{Email: email}

		var ue utils.Error
		if errors.As(err, &ue) {
			switch ue.Kind {
			case utils.InvalidData:
				if errs, ok := ue.Details.(validation.Errors); ok {
					params.EmailErr = errs["Email"]
				} else {
					me.logger.Warn("expected validation error", "found", err)
				}
			case utils.EmailNotFound:
				params.EmailErr = "Invalid email address"
			}

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

	session, err := me.authService.VerifyOtp(services.VerifyOtpParams{
		OtpID:    otpID,
		Otp:      otp,
		Platform: platform,
		OS:       os,
	})
	if err != nil {
		params := components.OtpFormParams{
			OtpID: otpID,
			Otp:   otp,
		}

		var ue utils.Error
		if errors.As(err, &ue) {
			switch ue.Kind {
			case utils.InvalidOtp:
				params.OtpErr = "Invalid code"
			}

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
	profile, err := me.profileService.GetProfile(getCurrentUserID(c))
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

// TODO: return an oob component to update the user block at the top of the sidebar
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

	if err := me.profileService.UpdateProfile(getCurrentUserID(c), services.UpdateProfileParams{
		Name:     name,
		Username: username,
		Email:    email,
		Bio:      bio,
	}); err != nil {
		var ue utils.Error
		if errors.As(err, &ue) {
			switch ue.Kind {
			case utils.InvalidData:
				if errs, ok := ue.Details.(validation.Errors); ok {
					params.NameErr = errs["Name"]
					params.UsernameErr = errs["Username"]
					params.EmailErr = errs["Email"]
					params.BioErr = errs["Bio"]
				} else {
					me.logger.Warn("expected validation error", "found", err)
				}
			case utils.UsernameConflict:
				params.UsernameErr = "Username is taken"
			case utils.EmailConflict:
				params.EmailErr = "Email is taken"
			}

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

	profileItems := make([]components.SearchProfileItem, 0, len(profiles))
	for _, p := range profiles {
		profileItems = append(profileItems, components.SearchProfileItem{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
		})
	}

	return render(c, components.SearchResult(components.SearchResultParams{
		Query:        query,
		HasMore:      len(profiles) == limit,
		ProfileItems: profileItems,
	}))
}
