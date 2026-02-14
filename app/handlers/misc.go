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
	"github.com/assaidy/h"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

// ApiError is the error type returned from API handlers
type ApiError struct {
	Kind        ErrorKind `json:"kind"`
	Description string    `json:"description"`
	Details     any       `json:"details,omitempty"`
}

func (me ApiError) Error() string {
	return fmt.Sprintf("%s: %v", me.Kind, me.Details)
}

type ErrorKind string

const (
	ErrInvalidJSON              ErrorKind = "InvalidJson"
	ErrInvalidData              ErrorKind = "InvalidData"
	ErrNotFound                 ErrorKind = "NotFound"
	ErrUsernameConflict         ErrorKind = "UsernameConflict"
	ErrEmailConflict            ErrorKind = "EmailConflict"
	ErrInternalFailure          ErrorKind = "InternalFailure"
	ErrEmailNotFound            ErrorKind = "EmailNotFound"
	ErrInvalidOTP               ErrorKind = "InvalidOtp"
	ErrUnauthorized             ErrorKind = "Unauthorized"
	ErrInvalidCursor            ErrorKind = "InvalidCursor"
	ErrInvalidEndpoint          ErrorKind = "InvalidEndpoint"
	ErrMethodNotAllowed         ErrorKind = "MethodNotAllowed"
	ErrWebscoketUpgradeRequired ErrorKind = "UpgradeRequired"
)

var errorDescriptions = map[ErrorKind]string{
	ErrInvalidJSON:              "The request body contains malformed or invalid JSON.",
	ErrInvalidData:              "The request data fails validation rules.",
	ErrNotFound:                 "The requested resource could not be found.",
	ErrUsernameConflict:         "Username already exists.",
	ErrEmailConflict:            "Email already exists.",
	ErrInternalFailure:          "An unexpected internal error occurred while processing the request.",
	ErrEmailNotFound:            "The provided email is not found.",
	ErrInvalidOTP:               "The provided otp is not invalid or expired.",
	ErrUnauthorized:             "Authentication is required or the provided credentials are invalid.",
	ErrInvalidCursor:            "The provided pagination cursor is malformed or invalid.",
	ErrInvalidEndpoint:          "The requested API endpoint does not exist or is malformed.",
	ErrMethodNotAllowed:         "The requested HTTP method is not allowed for this endpoint.",
	ErrWebscoketUpgradeRequired: "Websocket upgrade is required for this endpoint.",
}

func NewApiError(kind ErrorKind, details any) ApiError {
	return ApiError{
		Kind:        kind,
		Description: errorDescriptions[kind],
		Details:     details,
	}
}

// serviceErrToApiErr converts service errors to ApiError
func serviceErrToApiErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, services.ErrValidation) {
		var validationErrs validation.Errors
		if errors.As(err, &validationErrs) {
			return NewApiError(ErrInvalidData, validationErrs)
		}
		return NewApiError(ErrInvalidData, err)
	}

	if errors.Is(err, services.ErrNotFound) {
		return NewApiError(ErrNotFound, nil)
	}

	if errors.Is(err, services.ErrUnauthorized) {
		return NewApiError(ErrUnauthorized, nil)
	}

	if errors.Is(err, services.ErrInvalidOTP) {
		return NewApiError(ErrInvalidOTP, nil)
	}

	if errors.Is(err, services.ErrEmailNotFound) {
		return NewApiError(ErrEmailNotFound, nil)
	}

	if errors.Is(err, services.ErrUsernameTaken) {
		return NewApiError(ErrUsernameConflict, nil)
	}

	if errors.Is(err, services.ErrEmailTaken) {
		return NewApiError(ErrEmailConflict, nil)
	}

	// If it's not a known service error, return as-is (likely internal error)
	return err
}

// This error handler is called after [WithErrorResolver] and all other handlers,
// so err is ApiError, and the status code was set
func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		if c.Response().StatusCode() == fiber.StatusInternalServerError {
			if _, ok := err.(ApiError); !ok {
				logger.Warn("expected an api error", "error", err)
			}
			return c.JSON(NewApiError(ErrInternalFailure, nil))
		}

		return c.JSON(err)
	}
}

// This ensures we always return ApiError, and the proper status code was set.
// This must be registered after (handled before) the logger middleware
func WithErrorResolver(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			var code int
			var fe *fiber.Error
			var ae ApiError

			if errors.As(err, &fe) {
				// Returned by fiber's router
				code = fe.Code
				switch fe.Code {
				case fiber.StatusNotFound:
					err = NewApiError(ErrInvalidEndpoint, nil)
				case fiber.StatusMethodNotAllowed:
					err = NewApiError(ErrMethodNotAllowed, nil)
				default:
					logger.Warn("unhandled fiber error", "error", err)
					code = fiber.StatusInternalServerError
					err = NewApiError(ErrInternalFailure, err)
				}
			} else if errors.As(err, &ae) {
				switch ae.Kind {
				case ErrInvalidJSON, ErrInvalidData, ErrInvalidCursor:
					code = fiber.StatusBadRequest
				case ErrNotFound, ErrEmailNotFound, ErrInvalidEndpoint:
					code = fiber.StatusNotFound
				case ErrEmailConflict, ErrUsernameConflict:
					code = fiber.StatusConflict
				case ErrInvalidOTP, ErrUnauthorized:
					code = fiber.StatusUnauthorized
				case ErrWebscoketUpgradeRequired:
					code = fiber.StatusUpgradeRequired
				case ErrMethodNotAllowed:
					code = fiber.StatusMethodNotAllowed
				case ErrInternalFailure:
					code = fiber.StatusInternalServerError
				default:
					code = fiber.StatusInternalServerError
					logger.Warn("unhandled api error", "error", err)
				}
			} else {
				code = fiber.StatusInternalServerError
				err = NewApiError(ErrInternalFailure, err)
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
			return serviceErrToApiErr(err)
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
			return serviceErrToApiErr(err)
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
		var ae ApiError
		if errors.As(err, &ae) && ae.Kind == ErrUnauthorized {
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
			return NewApiError(ErrInvalidEndpoint, nil)
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
