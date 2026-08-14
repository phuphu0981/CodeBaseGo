package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codebasego/internal/platform/config"
)

// Client wraps goredis.Client for dependency injection.
type Client struct {
	*goredis.Client
	logger zerolog.Logger
}

// NewClient initializes a Redis client if enabled in config. Returns nil if disabled.
func NewClient(cfg *config.Config, log zerolog.Logger) (*Client, error) {
	if !cfg.Redis.Enabled {
		log.Info().Msg("redis is disabled in config; using in-memory components")
		return nil, nil
	}

	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}

	log.Info().Str("addr", addr).Msg("connected to redis server successfully")
	return &Client{
		Client: rdb,
		logger: log.With().Str("component", "redis").Logger(),
	}, nil
}
