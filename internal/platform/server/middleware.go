package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"codebasego/internal/common"
)

const HeaderRequestID = "X-Request-ID"

// requestID returns a middleware that assigns or extracts a unique request ID (Trace ID),
// attaches it to response headers, gin context, and request context.
func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(HeaderRequestID)
		if reqID == "" {
			reqID = c.GetHeader("X-Correlation-ID")
		}
		if reqID == "" {
			reqID = c.GetHeader("X-Trace-ID")
		}
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// Attach to HTTP response header
		c.Header(HeaderRequestID, reqID)
		c.Set("request_id", reqID)

		// Inject into standard Go context
		ctx := common.WithTraceID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// securityHeaders attaches standard security HTTP headers to protect against common web vulnerabilities.
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
