package common

import (
	"testing"
)

type samplePayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestDecodeEventPayload(t *testing.T) {
	t.Run("decode from struct", func(t *testing.T) {
		input := samplePayload{ID: "123", Name: "Alice"}
		var target samplePayload
		if err := DecodeEventPayload(input, &target); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target.ID != "123" || target.Name != "Alice" {
			t.Fatalf("mismatched payload: %+v", target)
		}
	})

	t.Run("decode from json string", func(t *testing.T) {
		input := `{"id":"456","name":"Bob"}`
		var target samplePayload
		if err := DecodeEventPayload(input, &target); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target.ID != "456" || target.Name != "Bob" {
			t.Fatalf("mismatched payload: %+v", target)
		}
	})

	t.Run("decode from nil", func(t *testing.T) {
		var target samplePayload
		if err := DecodeEventPayload(nil, &target); err != nil {
			t.Fatalf("unexpected error on nil: %v", err)
		}
	})
}
