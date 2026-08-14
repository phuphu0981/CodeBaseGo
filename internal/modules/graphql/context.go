package graphql

import (
	"context"

	"github.com/gin-gonic/gin"

	"codebasego/internal/common"
)

type contextKey string

const (
	userIDCtxKey contextKey = "user_id"
	emailCtxKey  contextKey = "email"
)

// GinContextToContextMiddleware transfers user_id and email from gin.Context into stdlib context.Context
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if val, exists := c.Get("user_id"); exists {
			if userID, ok := val.(string); ok {
				ctx = context.WithValue(ctx, userIDCtxKey, userID)
			}
		}

		if val, exists := c.Get("email"); exists {
			if email, ok := val.(string); ok {
				ctx = context.WithValue(ctx, emailCtxKey, email)
			}
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// GetUserIDFromContext retrieves user_id from GraphQL context. Returns "" if unauthenticated.
func GetUserIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(userIDCtxKey).(string); ok {
		return val
	}
	return ""
}

// GetEmailFromContext retrieves email from GraphQL context.
func GetEmailFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(emailCtxKey).(string); ok {
		return val
	}
	return ""
}

// RequireAuth extracts user_id from context or returns ErrUnauthorized.
func RequireAuth(ctx context.Context) (string, error) {
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		return "", common.ErrUnauthorized
	}
	return userID, nil
}
