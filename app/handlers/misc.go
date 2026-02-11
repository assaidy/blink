package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/gg"
	"github.com/gofiber/fiber/v2"
)

// This error handler is called after [WithErrorResolver] and all other handlers,
// so err is utils.Error, and the status code was set
func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		if c.Response().StatusCode() == fiber.StatusInternalServerError {
			if _, ok := err.(utils.Error); !ok {
				logger.Warn("expected a utils error", "error", err)
			}
			return c.JSON(utils.NewError(utils.InternalFailure, nil))
		}

		return c.JSON(err)
	}
}

// This ensures we always return utils.Error, and the proper status code was set.
// This must be registered after (handled before) the logger middleware
func WithErrorResolver(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			var code int
			var fe *fiber.Error
			var ue utils.Error

			if errors.As(err, &fe) {
				code = fe.Code
				switch fe.Code {
				case fiber.StatusNotFound:
					err = utils.NewError(utils.InvalidEndpoint, nil)
				case fiber.StatusMethodNotAllowed:
					err = utils.NewError(utils.MethodNotAllowed, nil)
				default:
					logger.Warn("unhandled fiber error", "error", err)
					code = fiber.StatusInternalServerError
					err = utils.NewError(utils.InternalFailure, err)
				}
			} else if errors.As(err, &ue) {
				switch ue.Kind {
				case utils.InvalidJson, utils.InvalidData, utils.InvalidCursor:
					code = fiber.StatusBadRequest
				case utils.NotFound, utils.EmailNotFound, utils.InvalidEndpoint:
					code = fiber.StatusNotFound
				case utils.EmailConflict, utils.UsernameConflict:
					code = fiber.StatusConflict
				case utils.InvalidOtp, utils.Unauthorized:
					code = fiber.StatusUnauthorized
				case utils.WebscoketUpgradeRequired:
					code = fiber.StatusUpgradeRequired
				case utils.MethodNotAllowed:
					code = fiber.StatusMethodNotAllowed
				case utils.InternalFailure:
					code = fiber.StatusInternalServerError
				default:
					code = fiber.StatusInternalServerError
					logger.Warn("unhandled utils error", "error", err)
				}
			} else {
				code = fiber.StatusInternalServerError
				err = utils.NewError(utils.InternalFailure, err)
			}

			c.Status(code)
			return err
		}

		return nil
	}
}

func WithLogging(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		took := time.Since(start)

		logger.Info("request handled",
			"took", took,
			"ip", c.IP(),
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"error", err,
		)

		return err
	}
}

func WithSessionToken(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, userID, err := authService.ValidateSessionToken(c.Cookies("session_token"))
		if err != nil {
			return err
		}

		c.Locals(currentSessionID, sessionID)
		c.Locals(currentUserID, userID)
		return c.Next()
	}
}

func WithSessionAndCsrfTokens(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, userID, err := authService.ValidateSessionAndCsrfTokens(c.Cookies("session_token"), c.Get("X-CSRF-Token"))
		if err != nil {
			return err
		}

		c.Locals(currentSessionID, sessionID)
		c.Locals(currentUserID, userID)
		return c.Next()
	}
}

const (
	currentSessionID = "Auth.SessionID"
	currentUserID    = "Auth.UserID"
)

func getCurrentSessionID(c *fiber.Ctx) string {
	return c.Locals(currentSessionID).(string)
}

func getCurrentUserID(c *fiber.Ctx) string {
	return c.Locals(currentUserID).(string)
}

func WithRedirectUnauthorizedToLogin(c *fiber.Ctx) error {
	if err := c.Next(); err != nil {
		var ue utils.Error
		if errors.As(err, &ue) && ue.Kind == utils.Unauthorized {
			return redirect(c, "/login")
		}
		return err
	}
	return nil
}

func WithForbiddenAsInvalidEndpoint(c *fiber.Ctx) error {
	if err := c.Next(); err != nil {
		var fe *fiber.Error
		if errors.As(err, &fe) && fe.Code == fiber.StatusForbidden {
			c.Status(fiber.StatusNotFound)
			return utils.NewError(utils.InvalidEndpoint, nil)
		}
		return err
	}
	return nil
}

func encodeCursor(v any) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func decodeCursor(decoded string, v any) error {
	bytes, err := base64.StdEncoding.DecodeString(decoded)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}
	if err := json.Unmarshal(bytes, v); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}
	return nil
}

type CursoredResponse[T any] struct {
	Items  []T    `json:"items"`
	Cursor string `json:"cursor,omitempty"`
}

func render(c *fiber.Ctx, component gg.Node) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return component.Render(c)
}

func redirect(c *fiber.Ctx, endpoint string) error {
	if c.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", endpoint)
	} else if c.Get("HX-Boosted") == "true" {
		c.Set("HX-Location", endpoint)
	} else {
		return c.Redirect(endpoint)
	}
	return nil
}

func extractPlatformAndOSFromUserAgent(userAgent string) (platform string, os string) {
	if userAgent == "" {
		return "Unknown", "Unknown"
	}

	ua := strings.ToLower(userAgent)

	os = "Unknown"
	for _, p := range osPatterns {
		if strings.Contains(ua, p.pattern) {
			os = p.name
			break
		}
	}

	platform = "Unknown"
	for _, p := range platformPatterns {
		if strings.Contains(ua, p.pattern) {
			platform = p.name
			break
		}
	}

	return platform, os
}

type pattern struct {
	pattern string
	name    string
}

var platformPatterns = []pattern{
	{"firefox", "Firefox"},
	{"chrome", "Chrome"},
	{"safari", "Safari"},
	{"edge", "Edge"},
	{"opera", "Opera"},
	{"brave", "Brave"},
}

var osPatterns = []pattern{
	{"windows nt 10.0", "Windows 10"},
	{"windows nt 6.3", "Windows 8.1"},
	{"windows nt 6.2", "Windows 8"},
	{"windows nt 6.1", "Windows 7"},
	{"windows", "Windows"},
	{"macintosh", "macOS"},
	{"mac os", "macOS"},
	{"linux", "Linux"},
	{"android", "Android"},
	{"iphone", "iOS"},
	{"ipad", "iOS"},
	{"ios", "iOS"},
}
