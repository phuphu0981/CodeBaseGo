package eventbus

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"codebasego/internal/common"
)

func TestMemoryEventBus(t *testing.T) {
	bus := NewMemoryEventBus(zerolog.Nop())

	t.Run("Publish and Subscribe success", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var received common.Event
		bus.Subscribe("user.created", func(ctx context.Context, event common.Event) error {
			received = event
			wg.Done()
			return nil
		})

		bus.Publish(context.Background(), common.Event{
			Name:    "user.created",
			Payload: "user-123",
		})

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			if received.Payload != "user-123" {
				t.Fatalf("expected payload user-123, got %v", received.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for event handler")
		}
	})

	t.Run("Handler Panic Recovery", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		bus.Subscribe("panic.event", func(ctx context.Context, event common.Event) error {
			defer wg.Done()
			panic("something went wrong in listener")
		})

		bus.Publish(context.Background(), common.Event{
			Name: "panic.event",
		})

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success: panic was recovered cleanly without crashing test
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for panic event handler")
		}
	})

	t.Run("Handler Returns Error", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		bus.Subscribe("error.event", func(ctx context.Context, event common.Event) error {
			defer wg.Done()
			return errors.New("handler error")
		})

		bus.Publish(context.Background(), common.Event{
			Name: "error.event",
		})

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for error event handler")
		}
	})
}
