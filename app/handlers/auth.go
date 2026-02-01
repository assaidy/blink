package handlers

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/blink/app/web/components"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	logger      *slog.Logger
	authService *services.AuthService
}

func NewAuthHandler(logger *slog.Logger, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		logger:      logger,
		authService: authService,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

func (me *AuthHandler) HandleApiRegister(c *fiber.Ctx) error {
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

func (me *AuthHandler) HandleApiRequestOtp(c *fiber.Ctx) error {
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

func extractPlatformAndOSFromUserAgent(userAgent string) (platform string, os string) {
	if userAgent == "" {
		return "Unknown", "Unknown"
	}

	ua := strings.ToLower(userAgent)

	os = "Unknown"
	osPatterns := map[string]string{
		"windows nt 10.0": "Windows 10",
		"windows nt 6.3":  "Windows 8.1",
		"windows nt 6.2":  "Windows 8",
		"windows nt 6.1":  "Windows 7",
		"windows":         "Windows",
		"macintosh":       "macOS",
		"mac os":          "macOS",
		"linux":           "Linux",
		"android":         "Android",
		"iphone":          "iOS",
		"ipad":            "iOS",
		"ios":             "iOS",
	}
	for pattern, osName := range osPatterns {
		if strings.Contains(ua, pattern) {
			os = osName
			break
		}
	}

	platform = "Unknown"
	platformPatterns := map[string]string{
		"firefox":   "Firefox",
		"chrome":    "Chrome",
		"safari":    "Safari",
		"edge":      "Edge",
		"opera":     "Opera",
		"brave":     "Brave",
		"iphone":    "iPhone",
		"ipad":      "iPad",
		"android":   "Android Device",
		"macintosh": "Mac",
		"windows":   "Windows PC",
		"linux":     "Linux PC",
	}
	for pattern, platformName := range platformPatterns {
		if strings.Contains(ua, pattern) {
			platform = platformName
			break
		}
	}

	return platform, os
}

func (me *AuthHandler) HandleApiVerifyOtp(c *fiber.Ctx) error {
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

const (
	currentSessionID = "Auth.SessionID"
	currentUserID    = "Auth.UserID"
)

func (me *AuthHandler) WithSessionToken(c *fiber.Ctx) error {
	sessionID, userID, err := me.authService.ValidateSessionToken(c.Cookies("session_token"))
	if err != nil {
		return err
	}

	c.Locals(currentSessionID, sessionID)
	c.Locals(currentUserID, userID)
	return c.Next()
}

func (me *AuthHandler) WithSessionAndCSRFTokens(c *fiber.Ctx) error {
	sessionID, userID, err := me.authService.ValidateSessionAndCsrfTokens(c.Cookies("session_token"), c.Get("X-CSRF-Token"))
	if err != nil {
		return err
	}

	c.Locals(currentSessionID, sessionID)
	c.Locals(currentUserID, userID)
	return c.Next()
}

func getCurrentSessionID(c *fiber.Ctx) string {
	return c.Locals(currentSessionID).(string)
}

func getCurrentUserID(c *fiber.Ctx) string {
	return c.Locals(currentUserID).(string)
}

func (me *AuthHandler) HandleApiLogout(c *fiber.Ctx) error {
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

func (me *AuthHandler) HandleApiGetActiveSessions(c *fiber.Ctx) error {
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

func (me *AuthHandler) HandleRegisterPage(c *fiber.Ctx) error {
	return render(c, components.RegisterPage())
}

func (me *AuthHandler) HandleRegister(c *fiber.Ctx) error {
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

		// TODO: return internal errors in a notification or a dedicated page
		return err
	}

	return redirect(c, "/login")
}

func (me *AuthHandler) HandleLoginPage(c *fiber.Ctx) error {
	return render(c, components.LoginPage())
}

func (me *AuthHandler) HandleLogin(c *fiber.Ctx) error {
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

func (me *AuthHandler) HandleVerifyOtp(c *fiber.Ctx) error {
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
