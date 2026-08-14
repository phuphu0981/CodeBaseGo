package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"codebasego/internal/common"
)

func TestIPRateLimiter(t *testing.T) {
	ctx := context.Background()
	limiter := newIPRateLimiter(2, 2)
	defer limiter.Stop()

	// Allow initial burst (2 tokens)
	if !limiter.Allow(ctx, "127.0.0.1") {
		t.Fatal("expected first request to be allowed")
	}
	if !limiter.Allow(ctx, "127.0.0.1") {
		t.Fatal("expected second request to be allowed")
	}

	// Third request exceeds burst
	if limiter.Allow(ctx, "127.0.0.1") {
		t.Fatal("expected third request to be blocked")
	}

	// Different IP should be allowed
	if !limiter.Allow(ctx, "192.168.1.1") {
		t.Fatal("expected request from different IP to be allowed")
	}
}

func TestIPRateLimiterCleanup(t *testing.T) {
	limiter := newIPRateLimiter(10, 10)
	defer limiter.Stop()
	limiter.clients.Store("old_ip", &clientBucket{
		tokens:     10,
		lastUpdate: time.Now().Add(-5 * time.Minute),
	})

	limiter.cleanupStale()

	if _, exists := limiter.clients.Load("old_ip"); exists {
		t.Fatal("expected old_ip to be cleaned up")
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(requestID())
	router.GET("/test", func(c *gin.Context) {
		reqID := common.GetTraceID(c.Request.Context())
		c.String(200, reqID)
	})

	t.Run("Generates Request ID when not provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		headerID := w.Header().Get(HeaderRequestID)
		if headerID == "" {
			t.Fatal("expected X-Request-ID header to be set")
		}
		if w.Body.String() != headerID {
			t.Fatalf("expected context trace ID %s, got %s", headerID, w.Body.String())
		}
	})

	t.Run("Preserves provided X-Request-ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set(HeaderRequestID, "custom-trace-12345")
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if w.Header().Get(HeaderRequestID) != "custom-trace-12345" {
			t.Fatalf("expected X-Request-ID custom-trace-12345, got %s", w.Header().Get(HeaderRequestID))
		}
		if w.Body.String() != "custom-trace-12345" {
			t.Fatalf("expected body custom-trace-12345, got %s", w.Body.String())
		}
	})
}
