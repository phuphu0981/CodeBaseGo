package common

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RouteRegistrar defines a standard interface for module route registration.
type RouteRegistrar interface {
	RegisterRoutes(rg *gin.RouterGroup)
}

// AuthMiddleware defines the typed authentication middleware for route protection.
type AuthMiddleware gin.HandlerFunc

// Migrator defines a standard interface for module database auto-migration.
type Migrator interface {
	AutoMigrate(db *gorm.DB) error
}

// BackgroundWorker defines a standard interface for module background workers with graceful shutdown support.
type BackgroundWorker interface {
	StartBackground(ctx context.Context, wg *sync.WaitGroup)
}

// Event represents a domain event in the system.
type Event struct {
	Name    string
	Payload any
	TraceID string
}

type contextKey string

const TraceIDContextKey contextKey = "trace_id"

// GetTraceID extracts the trace/request ID from context, or returns empty string if not found.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(TraceIDContextKey).(string); ok {
		return v
	}
	return ""
}

// WithTraceID injects a trace/request ID into the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDContextKey, traceID)
}

// EventHandler processes a domain event.
type EventHandler func(ctx context.Context, event Event) error

// EventBus provides publish-subscribe messaging for decoupled domain communication.
type EventBus interface {
	Subscribe(eventName string, handler EventHandler)
	Publish(ctx context.Context, event Event)
}

// DecodeEventPayload safely decodes a domain event payload into a target typed struct.
func DecodeEventPayload(payload any, target any) error {
	if payload == nil {
		return nil
	}
	switch p := payload.(type) {
	case string:
		return json.Unmarshal([]byte(p), target)
	case []byte:
		return json.Unmarshal(p, target)
	default:
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	}
}

