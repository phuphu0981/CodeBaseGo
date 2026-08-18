package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"admin.test+tag@domain.co.uk", true},
		{"invalid-email", false},
		{"@no-user.com", false},
		{"no-domain@", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsValidEmail(tt.email)
		if got != tt.valid {
			t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.valid)
		}
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Allowed role passes", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", "admin")
			c.Next()
		})
		r.GET("/admin-only", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Disallowed role is forbidden", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", "user")
			c.Next()
		})
		r.GET("/admin-only", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("Missing role is forbidden", func(t *testing.T) {
		r := gin.New()
		r.GET("/admin-only", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}
