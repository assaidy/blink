package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/assaidy/blink/app/utils"
	"github.com/gofiber/fiber/v2"
)

// this error hadnler is called after [WithErrorResolver] and all the other handlers,
// so err is utils.Error, and status code was set
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

// this ensures we always return utils.Error, and proper status code was set.
// this must be registered (handled) before the logger middleware
func WithErrorResolver(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			var code int
			var fe *fiber.Error
			var ue utils.Error

			if errors.As(err, &fe) {
				code = fe.Code
				switch fe.Code {
				// forbidden is returned from the filesystem middleware when triying to access an invalid resource
				case fiber.StatusForbidden, fiber.StatusNotFound:
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
				case utils.NotFound, utils.ClientNotFound, utils.EmailNotFound, utils.InvalidEndpoint:
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
