package page

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

type handlerMockPageRepo struct{}

func (m *handlerMockPageRepo) FindAll(ctx context.Context, query common.PaginationQuery) ([]Page, int, error) {
	pages := []Page{
		{ID: "page-1", Slug: "home", Title: "Home Page"},
	}
	return pages, 1, nil
}

func (m *handlerMockPageRepo) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]Page, common.CursorMeta, error) {
	pages := []Page{
		{ID: "page-1", Slug: "home", Title: "Home Page"},
	}
	return pages, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *handlerMockPageRepo) FindByID(ctx context.Context, id string) (*Page, error) {
	if id == "not-found" {
		return nil, common.ErrNotFound
	}
	return &Page{ID: id, Slug: "home", Title: "Home Page"}, nil
}

func (m *handlerMockPageRepo) FindBySlug(ctx context.Context, slug string) (*Page, error) {
	if slug == "not-found" {
		return nil, common.ErrNotFound
	}
	return &Page{ID: "page-1", Slug: slug, Title: "About Us"}, nil
}

func (m *handlerMockPageRepo) GetPublishedSlugs(ctx context.Context) ([]string, error) {
	return []string{"home", "about-us"}, nil
}

func (m *handlerMockPageRepo) Create(ctx context.Context, entity *Page) error {
	if entity.Slug == "existing" {
		return common.ErrConflict
	}
	return nil
}

func (m *handlerMockPageRepo) Update(ctx context.Context, entity *Page) error {
	if entity.ID == "not-found" {
		return common.ErrNotFound
	}
	return nil
}

func (m *handlerMockPageRepo) Delete(ctx context.Context, id string) error {
	if id == "not-found" {
		return common.ErrNotFound
	}
	return nil
}

func setupPageRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/pages/slugs", handler.GetSlugs)
	r.GET("/pages/by-slug", handler.GetBySlug)
	r.GET("/pages", handler.List)
	r.GET("/pages/:id", handler.GetByID)
	r.POST("/pages", handler.Create)
	r.PUT("/pages/:id", handler.Update)
	r.DELETE("/pages/:id", handler.Delete)
	return r
}

func TestPageHandler(t *testing.T) {
	svc := NewService(&handlerMockPageRepo{})
	h := NewHandler(svc)
	router := setupPageRouter(h)

	t.Run("List Pages", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/pages?page=1&per_page=10", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Published Slugs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/pages/slugs", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Page by ID - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/pages/page-1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Page by Slug - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/pages/by-slug?slug=about-us", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Page by Slug - Missing Slug", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/pages/by-slug", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("Create Page - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := CreatePageRequest{
			Slug:  "services",
			Title: "Our Services",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/pages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d", w.Code)
		}
	})

	t.Run("Update Page - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		newTitle := "Updated Title"
		payload := UpdatePageRequest{
			Title: &newTitle,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPut, "/pages/page-1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Delete Page - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/pages/page-1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})
}
