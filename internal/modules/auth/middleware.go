package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"codebasego/internal/platform/response"
)

// AuthMiddleware returns a Gin middleware that requires valid JWT Bearer tokens.
func (m *Module) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := m.service.ValidateToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT if present, but does not abort if missing.
func (m *Module) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				claims, err := m.service.ValidateToken(parts[1])
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("email", claims.Email)
				}
			}
		}
		c.Next()
	}
}
