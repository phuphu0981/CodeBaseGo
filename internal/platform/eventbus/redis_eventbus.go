package eventbus

import (
	"context"
	"encoding/json"
	"sync"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codebasego/internal/common"
	"codebasego/internal/platform/config"
	"codebasego/internal/platform/redis"
)

// RedisEventBus implements a distributed event bus backed by Redis Pub/Sub.
type RedisEventBus struct {
	client      *goredis.Client
	subscribers map[string][]common.EventHandler
	mu          sync.RWMutex
	logger      zerolog.Logger
}

// NewRedisEventBus creates a new RedisEventBus instance.
func NewRedisEventBus(redisClient *redis.Client, log zerolog.Logger) *RedisEventBus {
	var rdb *goredis.Client
	if redisClient != nil {
		rdb = redisClient.Client
	}
	return &RedisEventBus{
		client:      rdb,
		subscribers: make(map[string][]common.EventHandler),
		logger:      log.With().Str("component", "redis_eventbus").Logger(),
	}
}

// Subscribe registers an event handler and listens to the corresponding Redis channel.
func (b *RedisEventBus) Subscribe(eventName string, handler common.EventHandler) {
	b.mu.Lock()
	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
	first := len(b.subscribers[eventName]) == 1
	b.mu.Unlock()

	b.logger.Debug().Str("event", eventName).Msg("event handler subscribed for redis channel")

	if first && b.client != nil {
		go b.listen(eventName)
	}
}

func (b *RedisEventBus) listen(eventName string) {
	pubsub := b.client.Subscribe(context.Background(), eventName)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var evt common.Event
		if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
			b.logger.Error().Err(err).Str("event", eventName).Msg("failed to unmarshal redis event payload")
			continue
		}

		b.mu.RLock()
		handlers := make([]common.EventHandler, len(b.subscribers[eventName]))
		copy(handlers, b.subscribers[eventName])
		b.mu.RUnlock()

		for _, h := range handlers {
			handlerFunc := h
			asyncCtx := common.WithTraceID(context.WithoutCancel(context.Background()), evt.TraceID)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						b.logger.Error().
							Str("event", eventName).
							Str("trace_id", evt.TraceID).
							Interface("panic", r).
							Msg("redis event handler panicked")
					}
				}()
				if err := handlerFunc(asyncCtx, evt); err != nil {
					b.logger.Error().
						Err(err).
						Str("event", eventName).
						Str("trace_id", evt.TraceID).
						Msg("redis event handler returned error")
				}
			}()
		}
	}
}

// Publish dispatches an event to the Redis Pub/Sub channel.
func (b *RedisEventBus) Publish(ctx context.Context, event common.Event) {
	if b.client == nil {
		b.logger.Warn().Str("event", event.Name).Msg("redis client is nil; skipping redis publish")
		return
	}

	if event.TraceID == "" {
		event.TraceID = common.GetTraceID(ctx)
	}

	data, err := json.Marshal(event)
	if err != nil {
		b.logger.Error().
			Err(err).
			Str("event", event.Name).
			Str("trace_id", event.TraceID).
			Msg("failed to marshal event for redis publish")
		return
	}

	b.logger.Info().
		Str("event", event.Name).
		Str("trace_id", event.TraceID).
		Msg("publishing event to redis channel")

	if err := b.client.Publish(ctx, event.Name, string(data)).Err(); err != nil {
		b.logger.Error().
			Err(err).
			Str("event", event.Name).
			Str("trace_id", event.TraceID).
			Msg("failed to publish event to redis channel")
	}
}

// NewEventBus provides a dual-mode EventBus factory (Redis or In-Memory fallback).
func NewEventBus(cfg *config.Config, rdb *redis.Client, log zerolog.Logger) common.EventBus {
	if cfg.Redis.Enabled && rdb != nil && rdb.Client != nil {
		log.Info().Msg("initializing Redis Distributed EventBus")
		return NewRedisEventBus(rdb, log)
	}
	log.Info().Msg("initializing In-Memory EventBus")
	return NewMemoryEventBus(log)
}
