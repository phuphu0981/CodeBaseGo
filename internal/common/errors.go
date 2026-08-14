package common

import (
	"errors"
	"fmt"
	"net/mail"
)

// AppError represents a structured application error.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying wrapped error cause.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithCause wraps an underlying root error cause into a copy of AppError.
func (e *AppError) WithCause(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// Is supports matching via errors.Is based on the application/HTTP error code.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
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

