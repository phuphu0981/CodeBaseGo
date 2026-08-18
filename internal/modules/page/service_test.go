package page

import (
	"context"
	"testing"

	"codebasego/internal/common"
)

type mockPageRepo struct {
	pages map[string]*Page
	slugs map[string]*Page
}

func newMockPageRepo() *mockPageRepo {
	return &mockPageRepo{
		pages: make(map[string]*Page),
		slugs: make(map[string]*Page),
	}
}

func (m *mockPageRepo) FindAll(ctx context.Context, query common.PaginationQuery) ([]Page, int, error) {
	var list []Page
	for _, v := range m.pages {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockPageRepo) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]Page, common.CursorMeta, error) {
	var list []Page
	for _, v := range m.pages {
		list = append(list, *v)
	}
	return list, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *mockPageRepo) FindByID(ctx context.Context, id string) (*Page, error) {
	if p, ok := m.pages[id]; ok {
		return p, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockPageRepo) FindBySlug(ctx context.Context, slug string) (*Page, error) {
	if p, ok := m.slugs[slug]; ok {
		return p, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockPageRepo) GetPublishedSlugs(ctx context.Context) ([]string, error) {
	var list []string
	for _, v := range m.pages {
		if v.Status == "published" {
			list = append(list, v.Slug)
		}
	}
	return list, nil
}

func (m *mockPageRepo) Create(ctx context.Context, entity *Page) error {
	if _, exists := m.slugs[entity.Slug]; exists {
		return common.ErrConflict
	}
	if entity.ID == "" {
		entity.ID = "page-1"
	}
	m.pages[entity.ID] = entity
	m.slugs[entity.Slug] = entity
	return nil
}

func (m *mockPageRepo) Update(ctx context.Context, entity *Page) error {
	if _, ok := m.pages[entity.ID]; !ok {
		return common.ErrNotFound
	}
	m.pages[entity.ID] = entity
	m.slugs[entity.Slug] = entity
	return nil
}

func (m *mockPageRepo) Delete(ctx context.Context, id string) error {
	if p, ok := m.pages[id]; ok {
		delete(m.slugs, p.Slug)
		delete(m.pages, id)
		return nil
	}
	return common.ErrNotFound
}

func TestPageService(t *testing.T) {
	repo := newMockPageRepo()
	svc := NewService(repo)
	ctx := context.Background()

	t.Run("Create Page with empty slug fails", func(t *testing.T) {
		req := &CreatePageRequest{
			Slug:  "",
			Title: "Test Page",
		}
		_, err := svc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty slug")
		}
	})

	t.Run("Create Page with empty title fails", func(t *testing.T) {
		req := &CreatePageRequest{
			Slug:  "about-us",
			Title: "   ",
		}
		_, err := svc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty title")
		}
	})

	t.Run("Create valid Page succeeds and normalizes slug", func(t *testing.T) {
		req := &CreatePageRequest{
			Slug:     "/About-Us/",
			Title:    "About Us",
			Status:   "published",
			Template: "landing",
			Content:  `{"sections":[{"type":"hero"}]}`,
		}
		res, err := svc.Create(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Slug != "about-us" {
			t.Fatalf("expected slug 'about-us', got %s", res.Slug)
		}
		if res.Template != "landing" {
			t.Fatalf("expected template 'landing', got %s", res.Template)
		}
	})

	t.Run("Get Page by Slug succeeds", func(t *testing.T) {
		res, err := svc.GetBySlug(ctx, "about-us")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Title != "About Us" {
			t.Fatalf("expected title 'About Us', got %s", res.Title)
		}
	})

	t.Run("Get published slugs succeeds", func(t *testing.T) {
		slugs, err := svc.GetPublishedSlugs(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(slugs) != 1 || slugs[0] != "about-us" {
			t.Fatalf("expected ['about-us'], got %v", slugs)
		}
	})

	t.Run("Update Page succeeds", func(t *testing.T) {
		newTitle := "Updated About Us"
		req := &UpdatePageRequest{
			Title: &newTitle,
		}
		updated, err := svc.Update(ctx, "page-1", req)
		if err != nil {
			t.Fatalf("unexpected error on update: %v", err)
		}
		if updated.Title != newTitle {
			t.Fatalf("expected title %s, got %s", newTitle, updated.Title)
		}
	})

	t.Run("Delete Page succeeds", func(t *testing.T) {
		err := svc.Delete(ctx, "page-1")
		if err != nil {
			t.Fatalf("unexpected error on delete: %v", err)
		}
		_, err = svc.GetByID(ctx, "page-1")
		if err != common.ErrNotFound {
			t.Fatalf("expected ErrNotFound after delete, got %v", err)
		}
	})
}
