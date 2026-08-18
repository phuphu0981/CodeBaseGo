package seo

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

type handlerMockSEORepo struct{}

func (m *handlerMockSEORepo) FindAll(ctx context.Context, query common.PaginationQuery) ([]SEO, int, error) {
	records := []SEO{
		{ID: "seo-1", Slug: "home", Title: "Home Page"},
	}
	return records, 1, nil
}

func (m *handlerMockSEORepo) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]SEO, common.CursorMeta, error) {
	records := []SEO{
		{ID: "seo-1", Slug: "home", Title: "Home Page"},
	}
	return records, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *handlerMockSEORepo) FindByID(ctx context.Context, id string) (*SEO, error) {
	if id == "not-found" {
		return nil, common.ErrNotFound
	}
	return &SEO{ID: id, Slug: "home", Title: "Home Page"}, nil
}

func (m *handlerMockSEORepo) FindBySlug(ctx context.Context, slug string) (*SEO, error) {
	if slug == "not-found" {
		return nil, common.ErrNotFound
	}
	return &SEO{ID: "seo-1", Slug: slug, Title: "Test Page Title"}, nil
}

func (m *handlerMockSEORepo) Create(ctx context.Context, entity *SEO) error {
	if entity.Slug == "existing" {
		return common.ErrConflict
	}
	return nil
}

func (m *handlerMockSEORepo) Update(ctx context.Context, entity *SEO) error {
	if entity.ID == "not-found" {
		return common.ErrNotFound
	}
	return nil
}

func (m *handlerMockSEORepo) Delete(ctx context.Context, id string) error {
	if id == "not-found" {
		return common.ErrNotFound
	}
	return nil
}

func setupSEORouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/seo/by-slug", handler.GetBySlug)
	r.GET("/seo", handler.List)
	r.GET("/seo/:id", handler.GetByID)
	r.POST("/seo", handler.Create)
	r.PUT("/seo/:id", handler.Update)
	r.DELETE("/seo/:id", handler.Delete)
	return r
}

func TestSEOHandler(t *testing.T) {
	svc := NewService(&handlerMockSEORepo{})
	h := NewHandler(svc)
	router := setupSEORouter(h)

	t.Run("List SEO records", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/seo?page=1&per_page=10", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get SEO by ID - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/seo/seo-1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get SEO by ID - Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/seo/not-found", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("Get SEO by Slug - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/seo/by-slug?slug=about-us", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get SEO by Slug - Missing Slug", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/seo/by-slug", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("Create SEO record - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := CreateSEORequest{
			Slug:  "new-page",
			Title: "New Page Title",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/seo", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d", w.Code)
		}
	})

	t.Run("Create SEO record - Conflict", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := CreateSEORequest{
			Slug:  "existing",
			Title: "Existing Page",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/seo", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409 Conflict, got %d", w.Code)
		}
	})

	t.Run("Update SEO record - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		newTitle := "Updated Title"
		payload := UpdateSEORequest{
			Title: &newTitle,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/seo/seo-1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Delete SEO record - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/seo/seo-1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})
}
