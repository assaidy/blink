package services

import (
	"errors"
	"fmt"
)

var (
	ErrValidation       = errors.New("validation error")
	ErrNotFound         = errors.New("resource not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidOtp       = errors.New("invalid otp")
	ErrUsernameConflict = errors.New("username conflict")
	ErrEmailConflict    = errors.New("email conflict")
)

func makeEventChannelForUser(event string) func(userID string) string {
	return func(userID string) string {
		return fmt.Sprintf("events:%s:%s", event, userID)
	}
}
