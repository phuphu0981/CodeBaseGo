package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"codebasego/internal/common"
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

		role := claims.Role
		if role == "" {
			role = "user"
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", role)

		// Sync with standard Go context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, common.UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, common.EmailContextKey, claims.Email)
		ctx = context.WithValue(ctx, common.RoleContextKey, role)
		c.Request = c.Request.WithContext(ctx)

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
					role := claims.Role
					if role == "" {
						role = "user"
					}
					c.Set("user_id", claims.UserID)
					c.Set("email", claims.Email)
					c.Set("role", role)

					ctx := c.Request.Context()
					ctx = context.WithValue(ctx, common.UserIDContextKey, claims.UserID)
					ctx = context.WithValue(ctx, common.EmailContextKey, claims.Email)
					ctx = context.WithValue(ctx, common.RoleContextKey, role)
					c.Request = c.Request.WithContext(ctx)
				}
			}
		}
		c.Next()
	}
}

// RequireRole returns a Gin middleware ensuring the authenticated user has one of the required roles.
func (m *Module) RequireRole(roles ...string) gin.HandlerFunc {
	return common.RequireRole(roles...)
}

