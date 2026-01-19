package utils

import "fmt"

// inspired by https://core.telegram.org/method/auth.sendCode

type ErrorKind string

type Error struct {
	Kind        ErrorKind `json:"kind"`
	Description string    `json:"description"`
	Details     any       `json:"details,omitempty"`
}

func (me Error) Error() string {
	return fmt.Sprintf("%s: %v", me.Kind, me.Details)
}

var (
	InvalidJson              ErrorKind = "InvalidJson"
	InvalidData              ErrorKind = "InvalidData"
	NotFound                 ErrorKind = "NotFound"
	UsernameConflict         ErrorKind = "UsernameConflict"
	EmailConflict            ErrorKind = "EmailConflict"
	InternalFailure          ErrorKind = "InternalFailure"
	ClientNotFound           ErrorKind = "ClientNotFound"
	EmailNotFound            ErrorKind = "EmailNotFound"
	InvalidOtp               ErrorKind = "InvalidOtp"
	Unauthorized             ErrorKind = "Unauthorized"
	InvalidCursor            ErrorKind = "InvalidCursor"
	InvalidEndpoint          ErrorKind = "InvalidEndpoint"
	MethodNotAllowed         ErrorKind = "MethodNotAllowed"
	WebscoketUpgradeRequired ErrorKind = "UpgradeRequired"
)

var ErrorDescriptions = map[ErrorKind]string{
	InvalidJson:              "The request body contains malformed or invalid JSON.",
	InvalidData:              "The request data fails validation rules.",
	NotFound:                 "The requested resource could not be found.",
	UsernameConflict:         "Username already exists.",
	EmailConflict:            "Email already exists.",
	InternalFailure:          "An unexpected internal error occurred while processing the request.",
	ClientNotFound:           "The provided client id is not found",
	EmailNotFound:            "The provided email is not found",
	InvalidOtp:               "The provided otp is not invalid or expired",
	Unauthorized:             "Authentication is required or the provided credentials are invalid",
	InvalidCursor:            "The provided pagination cursor is malformed or invalid.",
	InvalidEndpoint:          "The requested API endpoint does not exist or is malformed.",
	MethodNotAllowed:         "The requested HTTP method is not allowed for this endpoint.",
	WebscoketUpgradeRequired: "Websocket upgrade is required for this endpoint",
}

func NewError(kind ErrorKind, details any) Error {
	return Error{
		Kind:        kind,
		Description: ErrorDescriptions[kind],
		Details:     details,
	}
}
