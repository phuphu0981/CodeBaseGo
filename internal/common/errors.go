package common

import (
	"fmt"
	"net/mail"
)

// AppError represents a structured application error.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

var (
	ErrNotFound     = NewAppError(404, "resource not found")
	ErrUnauthorized = NewAppError(401, "unauthorized")
	ErrForbidden    = NewAppError(403, "forbidden")
	ErrBadRequest   = NewAppError(400, "bad request")
	ErrConflict     = NewAppError(409, "resource already exists")
	ErrInternal     = NewAppError(500, "internal server error")
)

// IsValidEmail checks whether an email address format is valid.
func IsValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

