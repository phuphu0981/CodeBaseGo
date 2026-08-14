package common

import (
	"errors"
	"fmt"
	"testing"
)

func TestAppError_Is(t *testing.T) {
	errA := NewAppError(404, "not found")
	errB := NewAppError(404, "item not found in database")
	errC := NewAppError(500, "internal error")

	if !errors.Is(errA, errB) {
		t.Errorf("expected errA (404) to match errB (404) via errors.Is")
	}

	if errors.Is(errA, errC) {
		t.Errorf("expected errA (404) NOT to match errC (500)")
	}

	// Test wrapped error
	wrapped := fmt.Errorf("service failed: %w", errA)
	if !errors.Is(wrapped, ErrNotFound) {
		t.Errorf("expected wrapped error to match ErrNotFound via errors.Is")
	}
}

func TestAppError_WithCauseAndUnwrap(t *testing.T) {
	rootErr := errors.New("db connection lost")
	appErr := ErrInternal.WithCause(rootErr)

	if !errors.Is(appErr, ErrInternal) {
		t.Errorf("expected appErr to match ErrInternal")
	}

	if !errors.Is(appErr, rootErr) {
		t.Errorf("expected appErr to match rootErr through unwrap chain")
	}

	if appErr.Error() != "[500] internal server error: db connection lost" {
		t.Errorf("unexpected error string: %s", appErr.Error())
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"admin.test+tag@domain.co", true},
		{"invalid-email", false},
		{"@emptyuser.com", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.email, func(t *testing.T) {
			if got := IsValidEmail(tc.email); got != tc.valid {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tc.email, got, tc.valid)
			}
		})
	}
}
