package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"codebasego/internal/common"
	"codebasego/internal/platform/config"
	platformRedis "codebasego/internal/platform/redis"
)

type Server struct {
	Engine      *gin.Engine
	Config      *config.Config
	Logger      zerolog.Logger
	rateLimiter RateLimiter
}

func NewServer(cfg *config.Config, log zerolog.Logger, rdb *platformRedis.Client) *Server {
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	if len(cfg.Server.TrustedProxies) > 0 {
		_ = engine.SetTrustedProxies(cfg.Server.TrustedProxies)
	} else {
		_ = engine.SetTrustedProxies(nil)
	}

	engine.Use(gin.Recovery())
	engine.Use(requestID())
	engine.Use(securityHeaders())
	engine.Use(requestLogger(log))
	engine.Use(cors(cfg.CORS))

	var limiter RateLimiter
	if cfg.RateLimit.Enabled {
		limiter = NewRateLimiter(cfg, rdb, log)
		if limiter != nil {
			engine.Use(rateLimiterMiddleware(limiter))
		}
	}

	return &Server{
		Engine:      engine,
		Config:      cfg,
		Logger:      log,
		rateLimiter: limiter,
	}
}

func (s *Server) Run() error {
	port := s.Config.Server.Port
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		s.Logger.Info().Str("addr", addr).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Fatal().Err(err).Msg("server failed to listen")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.Logger.Info().Msg("shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.rateLimiter != nil {
		s.rateLimiter.Stop()
	}

	if err := srv.Shutdown(ctx); err != nil {
		s.Logger.Error().Err(err).Msg("server forced to shutdown")
		return err
	}

	s.Logger.Info().Msg("server exited cleanly")
	return nil
}

func requestLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		reqID := common.GetTraceID(c.Request.Context())
		if reqID == "" {
			reqID = c.GetString("request_id")
		}
		log.Info().
			Str("request_id", reqID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("request")
	}
}

func cors(cfg config.CORSConfig) gin.HandlerFunc {
	allowMethods := strings.Join(cfg.AllowMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowHeaders, ", ")
	allowCredentials := strconv.FormatBool(cfg.AllowCredentials)
	maxAge := strconv.Itoa(cfg.MaxAge)

	// Check if wildcard is used (allow all origins)
	isWildcard := len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*"

	// Build a set of allowed origins for fast lookup
	originSet := make(map[string]bool, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		originSet[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if isWildcard {
			if cfg.AllowCredentials && origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
			}
		} else if originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else {
			// Origin not allowed — skip CORS headers
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(403)
				return
			}
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Methods", allowMethods)
		c.Header("Access-Control-Allow-Headers", allowHeaders)
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", allowCredentials)
		c.Header("Access-Control-Max-Age", maxAge)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

