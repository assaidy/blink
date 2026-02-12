package services

import "errors"

var (
	ErrValidation    = errors.New("validation error")
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidOTP    = errors.New("invalid otp")
	ErrEmailNotFound = errors.New("email not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrEmailTaken    = errors.New("email already taken")
)
