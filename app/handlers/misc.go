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
	"github.com/assaidy/hyper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type apiError struct {
	Kind        errorKind `json:"kind"`
	Description string    `json:"description"`
	Details     any       `json:"details,omitempty"`
}

func newApiError(kind errorKind, details any) apiError {
	return apiError{
		Kind:        kind,
		Description: errorDescriptions[kind],
		Details:     details,
	}
}

func (me apiError) Error() string {
	return fmt.Sprintf("%s: %v", me.Kind, me.Details)
}

type errorKind string

const (
	errInvalidJSON              errorKind = "InvalidJson"
	errInvalidData              errorKind = "InvalidData"
	errNotFound                 errorKind = "NotFound"
	errUsernameConflict         errorKind = "UsernameConflict"
	errEmailConflict            errorKind = "EmailConflict"
	errInternalFailure          errorKind = "InternalFailure"
	errInvalidOtp               errorKind = "InvalidOtp"
	errUnauthorized             errorKind = "Unauthorized"
	errInvalidCursor            errorKind = "InvalidCursor"
	errInvalidEndpoint          errorKind = "InvalidEndpoint"
	errMethodNotAllowed         errorKind = "MethodNotAllowed"
	errWebscoketUpgradeRequired errorKind = "UpgradeRequired"
)

var errorDescriptions = map[errorKind]string{
	errInvalidJSON:              "The request body contains malformed or invalid JSON.",
	errInvalidData:              "The request data fails validation rules.",
	errNotFound:                 "The requested resource could not be found.",
	errUsernameConflict:         "Username already exists.",
	errEmailConflict:            "Email already exists.",
	errInternalFailure:          "An unexpected internal error occurred while processing the request.",
	errInvalidOtp:               "The provided otp is not invalid or expired.",
	errUnauthorized:             "Authentication is required or the provided credentials are invalid.",
	errInvalidCursor:            "The provided pagination cursor is malformed or invalid.",
	errInvalidEndpoint:          "The requested API endpoint does not exist or is malformed.",
	errMethodNotAllowed:         "The requested HTTP method is not allowed for this endpoint.",
	errWebscoketUpgradeRequired: "Websocket upgrade is required for this endpoint.",
}

func serviceErrToApiErr(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, services.ErrValidation):
		var validationErrs validation.Errors
		if errors.As(err, &validationErrs) {
			return newApiError(errInvalidData, validationErrs)
		}
		return newApiError(errInvalidData, err)
	case errors.Is(err, services.ErrNotFound):
		return newApiError(errNotFound, nil)
	case errors.Is(err, services.ErrUnauthorized):
		return newApiError(errUnauthorized, nil)
	case errors.Is(err, services.ErrInvalidOtp):
		return newApiError(errInvalidOtp, nil)
	case errors.Is(err, services.ErrUsernameConflict):
		return newApiError(errUsernameConflict, nil)
	case errors.Is(err, services.ErrEmailConflict):
		return newApiError(errEmailConflict, nil)
	}

	// If it's not a known service error, return as-is (likely internal error)
	return err
}

// ErrorHandler returns a Fiber error handler that returns [apiError] as JSON.
// It is called after [WithErrorResolver], so the status code is already set.
func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		if c.Response().StatusCode() == fiber.StatusInternalServerError {
			if _, ok := err.(apiError); !ok {
				logger.Warn("expected an api error", "error", err)
			}
			return c.JSON(newApiError(errInternalFailure, nil))
		}

		return c.JSON(err)
	}
}

// WithErrorResolver is a middleware that converts errors to apiError
// and sets the appropriate HTTP status code.
// It must be registered after (handled before) the logger middleware.
func WithErrorResolver(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			var code int
			var fe *fiber.Error
			var ae apiError

			if errors.As(err, &fe) {
				// Returned by fiber's router
				code = fe.Code
				switch fe.Code {
				case fiber.StatusNotFound:
					err = newApiError(errInvalidEndpoint, nil)
				case fiber.StatusMethodNotAllowed:
					err = newApiError(errMethodNotAllowed, nil)
				default:
					logger.Warn("unhandled fiber error", "error", err)
					code = fiber.StatusInternalServerError
					err = newApiError(errInternalFailure, err)
				}
			} else if errors.As(err, &ae) {
				// Returned by services
				switch ae.Kind {
				case errInvalidJSON, errInvalidData, errInvalidCursor:
					code = fiber.StatusBadRequest
				case errNotFound, errInvalidEndpoint:
					code = fiber.StatusNotFound
				case errEmailConflict, errUsernameConflict:
					code = fiber.StatusConflict
				case errInvalidOtp, errUnauthorized:
					code = fiber.StatusUnauthorized
				case errWebscoketUpgradeRequired:
					code = fiber.StatusUpgradeRequired
				case errMethodNotAllowed:
					code = fiber.StatusMethodNotAllowed
				case errInternalFailure:
					code = fiber.StatusInternalServerError
				default:
					code = fiber.StatusInternalServerError
					logger.Warn("unhandled api error", "error", err)
				}
			} else {
				// General error returned from anywhere.
				// I don't give specific error type for general errors (e.g. DB failiur, encodin/decoding errors, ...etc).
				// They are all handled here.
				code = fiber.StatusInternalServerError
				err = newApiError(errInternalFailure, err)
			}

			c.Status(code)
			return err
		}

		return nil
	}
}

// WithLogging is a middleware that logs request details (method, path, status, duration, IP).
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

const (
	currentSessionID = "Auth.SessionID"
	currentUserID    = "Auth.UserID"
)

// WithSessionTokenCookie is a middleware that validates the session token from cookies
// and stores the session ID and user ID in context locals.
func WithSessionTokenCookie(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, userID, err := authService.ValidateSessionToken(c.Cookies("session_token"))
		if err != nil {
			return serviceErrToApiErr(err)
		}

		c.Locals(currentSessionID, sessionID)
		c.Locals(currentUserID, userID)
		return c.Next()
	}
}

// getCurrentSessionID returns the session ID from the request context.
// It must be called after [WithSessionTokenCookie] middleware.
func getCurrentSessionID(c *fiber.Ctx) string {
	return c.Locals(currentSessionID).(string)
}

// getCurrentUserID returns the user ID from the request context.
// It must be called after [WithSessionTokenCookie] middleware.
func getCurrentUserID(c *fiber.Ctx) string {
	return c.Locals(currentUserID).(string)
}

// WithCsrfTokenHeader is a middleware that validates the CSRF token from the X-CSRF-Token header.
// It must be used after passing through [WithSessionTokenCookie]
func WithCsrfTokenHeader(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := authService.ValidateCsrfToken(getCurrentSessionID(c), c.Get("X-CSRF-Token")); err != nil {
			return serviceErrToApiErr(err)
		}
		return c.Next()
	}
}

// WithCsrfTokenQuery is a middleware that validates the CSRF token from the csrf_token URL query.
// It must be used after passing through [WithSessionTokenCookie]
func WithCsrfTokenQuery(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := authService.ValidateCsrfToken(getCurrentSessionID(c), c.Query("csrf_token")); err != nil {
			return serviceErrToApiErr(err)
		}
		return c.Next()
	}
}

// WithWebsocket is a middleware that checks if the request is a WebSocket upgrade.
// If not, it returns an error indicating that a WebSocket upgrade is required.
func WithWebsocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return newApiError(errWebscoketUpgradeRequired, nil)
}

// WithRedirectUnauthorizedToLogin is a middleware that redirects unauthorized users to the login page.
func WithRedirectUnauthorizedToLogin(c *fiber.Ctx) error {
	if err := c.Next(); err != nil {
		var ae apiError
		if errors.As(err, &ae) && ae.Kind == errUnauthorized {
			return redirect(c, "/login")
		}
		return err
	}
	return nil
}

// WithForbiddenAsInvalidEndpoint is a middleware that treats 403 Forbidden as 404 Not Found.
// This prevents information leakage about the existence of protected resources.
func WithForbiddenAsInvalidEndpoint(c *fiber.Ctx) error {
	if err := c.Next(); err != nil {
		var fe *fiber.Error
		if errors.As(err, &fe) && fe.Code == fiber.StatusForbidden {
			c.Status(fiber.StatusNotFound)
			return newApiError(errInvalidEndpoint, nil)
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

type cursoredResponse[T any] struct {
	Items  []T    `json:"items"`
	Cursor string `json:"cursor,omitempty"`
}

func render(c *fiber.Ctx, component h.Node) error {
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

const (
	partnerPresenceChanged   = "ChatPartnerPresenceChanged"
	chatWasDeleted           = "ChatWasDeleted"
	userProfileWasUpdated    = "UserProfileWasUpdated"
	partnerProfileWasUpdated = "PartnerProfileWasUpdated"
	partnerProfileWasDeleted = "PartnerProfileWasDeleted"
	sendMessage              = "SendMessage"
	messageWasSent           = "MessageWasSent"
	incommingMessage         = "IncommingMessage"
	userMessagesWereRead     = "UserMessagesWereRead"
	partnerMessagesWereRead  = "PartnerMessagesWereRead"
)

type websocketMessage struct {
	Kind              string    `json:"kind"`
	UserID            string    `json:"userID,omitempty"`
	PartnerID         string    `json:"partnerID,omitempty"`
	IsOnline          bool      `json:"isOnline,omitempty"`
	ReadMessageIDs    []string  `json:"readMessageIDs,omitempty"`
	ReadMessagesCount int       `json:"readMessagesCount,omitempty"`
	Name              string    `json:"name,omitempty"`
	Username          string    `json:"username,omitempty"`
	Email             string    `json:"email,omitempty"`
	Bio               string    `json:"bio,omitempty"`
	MessageID         string    `json:"messageID,omitempty"`
	ClientMessageID   int       `json:"clientMessageID,omitempty"`
	Content           string    `json:"content,omitempty"`
	Timestamp         time.Time `json:"timestamp,omitempty"`
}

func extractPlatformAndOSFromUserAgent(ua string) (platform string, os string) {
	if ua == "" {
		return "Unknown", "Unknown"
	}
	ua = strings.ToLower(ua)

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
	// Order is critical
	{"firefox", "Firefox"},
	{"chrome", "Chrome"},
	{"safari", "Safari"},
	{"edge", "Edge"},
	{"opera", "Opera"},
	{"brave", "Brave"},
}

var osPatterns = []pattern{
	// Order is critical
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
