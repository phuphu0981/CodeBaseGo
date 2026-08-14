package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"codebasego/internal/platform/config"
)

func setupAuthRouter(handler *Handler, module *Module) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	module.RegisterRoutes(v1)
	return r
}

func TestAuthHandler(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "test-secret-key-for-handler-tests",
			AccessExpireMinute: 15,
			RefreshExpireDay:   7,
		},
	}
	userSvc := newMockUserService()
	refreshRepo := newMockRefreshTokenRepo()
	svc := NewService(cfg, userSvc, refreshRepo, nil)
	h := NewHandler(svc)
	mod := NewModule(h, svc)
	router := setupAuthRouter(h, mod)

	t.Run("Register - Success", func(t *testing.T) {
		body, _ := json.Marshal(RegisterRequest{
			Email:    "newuser@example.com",
			Password: "password123",
			Name:     "New User",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d", w.Code)
		}
	})

	t.Run("Register - Duplicate Email Conflict", func(t *testing.T) {
		body, _ := json.Marshal(RegisterRequest{
			Email:    "newuser@example.com",
			Password: "password123",
			Name:     "New User",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict, got %d", w.Code)
		}
	})

	t.Run("Login - Success", func(t *testing.T) {
		body, _ := json.Marshal(LoginRequest{
			Email:    "newuser@example.com",
			Password: "password123",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", w.Code)
		}
	})

	t.Run("Login - Invalid Credentials", func(t *testing.T) {
		body, _ := json.Marshal(LoginRequest{
			Email:    "newuser@example.com",
			Password: "wrongpassword",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d", w.Code)
		}
	})

	t.Run("Refresh Token Flow", func(t *testing.T) {
		pair, err := svc.GenerateTokenPair(t.Context(), "user-1", "newuser@example.com")
		if err != nil {
			t.Fatalf("failed to generate token pair: %v", err)
		}

		body, _ := json.Marshal(RefreshTokenRequest{
			RefreshToken: pair.RefreshToken,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", w.Code)
		}
	})

	t.Run("Logout-All - Unauthorized without bearer token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout-all", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d", w.Code)
		}
	})

	t.Run("Logout-All - Success with valid bearer token", func(t *testing.T) {
		pair, _ := svc.GenerateTokenPair(t.Context(), "user-1", "newuser@example.com")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout-all", nil)
		req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", w.Code)
		}
	})
}
