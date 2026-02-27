package services

import "errors"

var (
	ErrValidation       = errors.New("validation error")
	ErrNotFound         = errors.New("resource not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidOTP       = errors.New("invalid otp")
	ErrUsernameConflict = errors.New("username conflict")
	ErrEmailConflict    = errors.New("email conflict")
)
