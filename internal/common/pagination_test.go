package common

import (
	"testing"
	"time"
)

func TestCursorEncodingDecoding(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	userID := "user-123-uuid"

	encoded := EncodeCursor(now, userID)
	if encoded == "" {
		t.Fatalf("expected non-empty encoded cursor")
	}

	decodedTime, decodedID, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %v", err)
	}

	if decodedID != userID {
		t.Errorf("expected userID %s, got %s", userID, decodedID)
	}

	if !decodedTime.Equal(now) {
		t.Errorf("expected time %v, got %v", now, decodedTime)
	}
}

func TestInvalidCursorDecoding(t *testing.T) {
	_, _, err := DecodeCursor("invalid-base64-!!!")
	if err == nil {
		t.Errorf("expected error decoding invalid base64")
	}

	// Base64 valid but missing '|'
	encodedNoPipe := "c29tZXN0cmluZw=="
	_, _, err = DecodeCursor(encodedNoPipe)
	if err == nil {
		t.Errorf("expected error decoding cursor without pipe delimiter")
	}
}
