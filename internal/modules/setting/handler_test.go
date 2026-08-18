package setting

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

type handlerMockSettingRepo struct{}

func (m *handlerMockSettingRepo) FindAll(ctx context.Context, query common.PaginationQuery, scope, pathPrefix string) ([]CoreConfig, int, error) {
	configs := []CoreConfig{
		{ID: "cfg-1", Scope: "default", ScopeID: "0", Path: PathBaseURL, Value: "http://localhost:3000"},
	}
	return configs, 1, nil
}

func (m *handlerMockSettingRepo) FindAllCursor(ctx context.Context, query common.CursorQuery, scope, pathPrefix string) ([]CoreConfig, common.CursorMeta, error) {
	configs := []CoreConfig{
		{ID: "cfg-1", Scope: "default", ScopeID: "0", Path: PathBaseURL, Value: "http://localhost:3000"},
	}
	return configs, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *handlerMockSettingRepo) FindByID(ctx context.Context, id string) (*CoreConfig, error) {
	if id == "not-found" {
		return nil, common.ErrNotFound
	}
	return &CoreConfig{ID: id, Scope: "default", ScopeID: "0", Path: PathBaseURL, Value: "http://localhost:3000"}, nil
}

func (m *handlerMockSettingRepo) FindByPath(ctx context.Context, scope, scopeID, path string) (*CoreConfig, error) {
	if path == "not-found" {
		return nil, common.ErrNotFound
	}
	return &CoreConfig{ID: "cfg-1", Scope: scope, ScopeID: scopeID, Path: path, Value: "http://localhost:3000"}, nil
}

func (m *handlerMockSettingRepo) FindByPrefix(ctx context.Context, scope, scopeID, prefix string) ([]CoreConfig, error) {
	return []CoreConfig{
		{ID: "cfg-1", Scope: scope, ScopeID: scopeID, Path: "web/public/title", Value: "My Store"},
	}, nil
}

func (m *handlerMockSettingRepo) Save(ctx context.Context, entity *CoreConfig) error {
	return nil
}

func (m *handlerMockSettingRepo) Delete(ctx context.Context, id string) error {
	if id == "not-found" {
		return common.ErrNotFound
	}
	return nil
}

func (m *handlerMockSettingRepo) DeleteByPath(ctx context.Context, scope, scopeID, path string) error {
	return nil
}

func setupSettingRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/settings/public", handler.GetPublic)
	r.GET("/settings/by-path", handler.GetByPath)
	r.GET("/settings", handler.List)
	r.GET("/settings/:id", handler.GetByID)
	r.POST("/settings", handler.Set)
	r.DELETE("/settings/:id", handler.Delete)
	return r
}

func TestSettingHandler(t *testing.T) {
	svc := NewService(&handlerMockSettingRepo{})
	h := NewHandler(svc)
	router := setupSettingRouter(h)

	t.Run("Get Public Configs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/settings/public", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Config by Path - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/settings/by-path?path="+PathBaseURL, nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Config by Path - Missing Path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/settings/by-path", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("List Settings", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/settings?page=1&per_page=10", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Set Config - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := SetConfigRequest{
			Path:  "custom/key",
			Value: "my-value",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/settings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Delete Config - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/settings/cfg-1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})
}
