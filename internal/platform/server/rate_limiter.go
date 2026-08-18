package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codebasego/internal/platform/config"
	platformRedis "codebasego/internal/platform/redis"
	"codebasego/internal/platform/response"
)

// RateLimiter interface defines rate limiting contracts.
type RateLimiter interface {
	Allow(ctx context.Context, key string) bool
	Stop()
}

// ==============================================================================
// In-Memory Token Bucket Rate Limiter
// ==============================================================================

type clientBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastUpdate time.Time
}

type ipRateLimiter struct {
	clients sync.Map
	rps     float64
	burst   float64
	stopCh  chan struct{}
	once    sync.Once
}

func newIPRateLimiter(rps, burst int) *ipRateLimiter {
	limiter := &ipRateLimiter{
		rps:    float64(rps),
		burst:  float64(burst),
		stopCh: make(chan struct{}),
	}
	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				limiter.cleanupStale()
			case <-limiter.stopCh:
				return
			}
		}
	}()
	return limiter
}

func (l *ipRateLimiter) Stop() {
	l.once.Do(func() {
		close(l.stopCh)
	})
}

func (l *ipRateLimiter) cleanupStale() {
	now := time.Now()
	l.clients.Range(func(key, value any) bool {
		if bucket, ok := value.(*clientBucket); ok {
			bucket.mu.Lock()
			if now.Sub(bucket.lastUpdate) > 3*time.Minute {
				l.clients.Delete(key)
			}
			bucket.mu.Unlock()
		}
		return true
	})
}

func (l *ipRateLimiter) Allow(ctx context.Context, ip string) bool {
	now := time.Now()

	for {
		val, loaded := l.clients.Load(ip)
		var bucket *clientBucket
		if !loaded {
			newBucket := &clientBucket{
				tokens:     l.burst,
				lastUpdate: now,
			}
			actual, _ := l.clients.LoadOrStore(ip, newBucket)
			bucket = actual.(*clientBucket)
		} else {
			bucket = val.(*clientBucket)
		}

		bucket.mu.Lock()
		if now.Sub(bucket.lastUpdate) > 3*time.Minute && loaded {
			bucket.mu.Unlock()
			l.clients.Delete(ip)
			continue
		}

		elapsed := now.Sub(bucket.lastUpdate).Seconds()
		bucket.tokens += elapsed * l.rps
		if bucket.tokens > l.burst {
			bucket.tokens = l.burst
		}
		bucket.lastUpdate = now

		allowed := bucket.tokens >= 1
		if allowed {
			bucket.tokens -= 1
		}
		bucket.mu.Unlock()
		return allowed
	}
}

// ==============================================================================
// Distributed Redis Rate Limiter
// ==============================================================================

const redisTokenBucketScript = `
local key = KEYS[1]
local burst = tonumber(ARGV[1])
local rps = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local data = redis.call('HMGET', key, 'tokens', 'last_update')
local tokens = tonumber(data[1])
local last_update = tonumber(data[2])

if not tokens then
    tokens = burst
    last_update = now
else
    local elapsed = math.max(0, now - last_update)
    tokens = math.min(burst, tokens + elapsed * rps)
    last_update = now
end

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call('HSET', key, 'tokens', tokens, 'last_update', last_update)
local expireSec = math.ceil(burst / math.max(1, rps)) + 60
redis.call('EXPIRE', key, expireSec)

return allowed
`

type redisRateLimiter struct {
	rdb      *platformRedis.Client
	fallback *ipRateLimiter
	rps      int
	burst    int
	script   *redis.Script
	logger   zerolog.Logger
}

func newRedisRateLimiter(rdb *platformRedis.Client, rps, burst int, log zerolog.Logger) *redisRateLimiter {
	return &redisRateLimiter{
		rdb:      rdb,
		fallback: newIPRateLimiter(rps, burst),
		rps:      rps,
		burst:    burst,
		script:   redis.NewScript(redisTokenBucketScript),
		logger:   log.With().Str("component", "redis_rate_limiter").Logger(),
	}
}

func (r *redisRateLimiter) Allow(ctx context.Context, ip string) bool {
	if r.rdb == nil || r.rdb.Client == nil {
		return r.fallback.Allow(ctx, ip)
	}

	key := fmt.Sprintf("ratelimit:%s", ip)
	now := float64(time.Now().UnixNano()) / 1e9

	res, err := r.script.Run(ctx, r.rdb.Client, []string{key}, r.burst, r.rps, now).Int()
	if err != nil {
		r.logger.Warn().Err(err).Msg("redis rate limit failed; falling back to in-memory")
		return r.fallback.Allow(ctx, ip)
	}

	return res == 1
}

func (r *redisRateLimiter) Stop() {
	if r.fallback != nil {
		r.fallback.Stop()
	}
}

// NewRateLimiter instantiates appropriate RateLimiter (Redis or In-Memory) based on configuration.
func NewRateLimiter(cfg *config.Config, rdb *platformRedis.Client, log zerolog.Logger) RateLimiter {
	if !cfg.RateLimit.Enabled {
		return nil
	}

	if cfg.Redis.Enabled && rdb != nil {
		log.Info().Msg("using distributed Redis rate limiter")
		return newRedisRateLimiter(rdb, cfg.RateLimit.RPS, cfg.RateLimit.Burst, log)
	}

	log.Info().Msg("using in-memory token bucket rate limiter")
	return newIPRateLimiter(cfg.RateLimit.RPS, cfg.RateLimit.Burst)
}

func rateLimiterMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !limiter.Allow(c.Request.Context(), clientIP) {
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded, please try again later")
			c.Abort()
			return
		}
		c.Next()
	}
}
