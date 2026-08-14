package eventbus

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"codebasego/internal/common"
)

// MemoryEventBus implements an in-memory thread-safe event bus.
type MemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]common.EventHandler
	logger      zerolog.Logger
}

// NewMemoryEventBus creates a new MemoryEventBus instance.
func NewMemoryEventBus(log zerolog.Logger) *MemoryEventBus {
	return &MemoryEventBus{
		subscribers: make(map[string][]common.EventHandler),
		logger:      log.With().Str("component", "eventbus").Logger(),
	}
}

// Subscribe registers a handler for a specific event name.
func (b *MemoryEventBus) Subscribe(eventName string, handler common.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
	b.logger.Debug().Str("event", eventName).Msg("event handler subscribed")
}

// Publish dispatches an event asynchronously to all registered handlers.
func (b *MemoryEventBus) Publish(ctx context.Context, event common.Event) {
	if event.TraceID == "" {
		event.TraceID = common.GetTraceID(ctx)
	}

	b.mu.RLock()
	handlers, exists := b.subscribers[event.Name]
	if !exists || len(handlers) == 0 {
		b.mu.RUnlock()
		b.logger.Debug().
			Str("event", event.Name).
			Str("trace_id", event.TraceID).
			Msg("no handlers registered for event")
		return
	}

	// Copy handlers slice to release read lock quickly
	handlersCopy := make([]common.EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	b.mu.RUnlock()

	b.logger.Info().
		Str("event", event.Name).
		Str("trace_id", event.TraceID).
		Int("subscribers", len(handlersCopy)).
		Msg("publishing event")

	for _, h := range handlersCopy {
		handlerFunc := h
		asyncCtx := common.WithTraceID(context.WithoutCancel(ctx), event.TraceID)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					b.logger.Error().
						Str("event", event.Name).
						Str("trace_id", event.TraceID).
						Interface("panic", r).
						Msg("event handler panicked")
				}
			}()

			if err := handlerFunc(asyncCtx, event); err != nil {
				b.logger.Error().
					Err(err).
					Str("event", event.Name).
					Str("trace_id", event.TraceID).
					Msg("event handler returned error")
			}
		}()
	}
}
