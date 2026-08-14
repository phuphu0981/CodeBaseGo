package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"codebasego/internal/common"
)

type handlerMockRepo struct{}

func (m *handlerMockRepo) FindAll(ctx context.Context, query common.PaginationQuery) ([]User, int, error) {
	users := []User{
		{ID: "user-1", Email: "user1@example.com", Name: "User 1"},
	}
	return users, 1, nil
}

func (m *handlerMockRepo) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]User, common.CursorMeta, error) {
	users := []User{
		{ID: "user-1", Email: "user1@example.com", Name: "User 1"},
	}
	return users, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *handlerMockRepo) FindByID(ctx context.Context, id string) (*User, error) {
	if id == "not-found" {
		return nil, common.ErrNotFound
	}
	return &User{ID: id, Email: "user@example.com", Name: "User"}, nil
}

func (m *handlerMockRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	if email == "existing@example.com" {
		return &User{ID: "other-user", Email: "existing@example.com"}, nil
	}
	return nil, common.ErrNotFound
}

func (m *handlerMockRepo) Create(ctx context.Context, entity *User) error {
	return nil
}

func (m *handlerMockRepo) Update(ctx context.Context, entity *User) error {
	if entity.Email == "existing@example.com" {
		return common.ErrConflict
	}
	if entity.ID == "not-found" {
		return common.ErrNotFound
	}
	return nil
}

func (m *handlerMockRepo) Delete(ctx context.Context, id string) error {
	if id == "not-found" {
		return common.ErrNotFound
	}
	return nil
}

func setupUserRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// Mock auth middleware setting user_id
		if uid := c.GetHeader("X-User-ID"); uid != "" {
			c.Set("user_id", uid)
		}
		c.Next()
	})
	r.GET("/users", handler.List)
	r.GET("/users/:id", handler.GetByID)
	r.PUT("/users/:id", handler.Update)
	r.DELETE("/users/:id", handler.Delete)
	return r
}

func TestUserHandler(t *testing.T) {
	svc := NewService(&handlerMockRepo{})
	h := NewHandler(svc)
	router := setupUserRouter(h)

	t.Run("List users", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/users?page=1&per_page=10", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get user by ID - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/users/user-1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get user by ID - Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/users/not-found", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("Update user - Forbidden when modifying different user", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(UpdateRequest{})
		req, _ := http.NewRequest(http.MethodPut, "/users/user-1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "different-user")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("Update user - Conflict when email taken", func(t *testing.T) {
		w := httptest.NewRecorder()
		email := "existing@example.com"
		body, _ := json.Marshal(UpdateRequest{Email: &email})
		req, _ := http.NewRequest(http.MethodPut, "/users/user-1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-1")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict, got %d", w.Code)
		}
	})

	t.Run("Delete user - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/users/user-1", nil)
		req.Header.Set("X-User-ID", "user-1")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})
}
