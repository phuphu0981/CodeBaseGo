package eventbus

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"codebasego/internal/common"
	"codebasego/internal/platform/config"
)

func TestNewEventBusFallback(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Enabled: false,
		},
	}
	log := zerolog.Nop()

	bus := NewEventBus(cfg, nil, log)
	if bus == nil {
		t.Fatal("expected non-nil EventBus instance")
	}

	received := make(chan bool, 1)
	bus.Subscribe("user.created", func(ctx context.Context, event common.Event) error {
		received <- true
		return nil
	})

	bus.Publish(context.Background(), common.Event{Name: "user.created", Payload: "test"})

	select {
	case <-received:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for event bus handler")
	}
}
