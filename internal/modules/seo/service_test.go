package seo

import (
	"context"
	"testing"

	"codebasego/internal/common"
)

type mockSEORepo struct {
	records map[string]*SEO
	slugs   map[string]*SEO
}

func newMockSEORepo() *mockSEORepo {
	return &mockSEORepo{
		records: make(map[string]*SEO),
		slugs:   make(map[string]*SEO),
	}
}

func (m *mockSEORepo) FindAll(ctx context.Context, query common.PaginationQuery) ([]SEO, int, error) {
	var list []SEO
	for _, v := range m.records {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockSEORepo) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]SEO, common.CursorMeta, error) {
	var list []SEO
	for _, v := range m.records {
		list = append(list, *v)
	}
	return list, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *mockSEORepo) FindByID(ctx context.Context, id string) (*SEO, error) {
	if rec, ok := m.records[id]; ok {
		return rec, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockSEORepo) FindBySlug(ctx context.Context, slug string) (*SEO, error) {
	if rec, ok := m.slugs[slug]; ok {
		return rec, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockSEORepo) Create(ctx context.Context, entity *SEO) error {
	if _, exists := m.slugs[entity.Slug]; exists {
		return common.ErrConflict
	}
	if entity.ID == "" {
		entity.ID = "seo-1"
	}
	m.records[entity.ID] = entity
	m.slugs[entity.Slug] = entity
	return nil
}

func (m *mockSEORepo) Update(ctx context.Context, entity *SEO) error {
	if _, ok := m.records[entity.ID]; !ok {
		return common.ErrNotFound
	}
	m.records[entity.ID] = entity
	m.slugs[entity.Slug] = entity
	return nil
}

func (m *mockSEORepo) Delete(ctx context.Context, id string) error {
	if rec, ok := m.records[id]; ok {
		delete(m.slugs, rec.Slug)
		delete(m.records, id)
		return nil
	}
	return common.ErrNotFound
}

func TestSEOService(t *testing.T) {
	repo := newMockSEORepo()
	svc := NewService(repo)
	ctx := context.Background()

	t.Run("Create SEO record with empty slug fails", func(t *testing.T) {
		req := &CreateSEORequest{
			Slug:  "",
			Title: "Test Title",
		}
		_, err := svc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty slug")
		}
	})

	t.Run("Create SEO record with empty title fails", func(t *testing.T) {
		req := &CreateSEORequest{
			Slug:  "about-us",
			Title: "   ",
		}
		_, err := svc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty title")
		}
	})

	t.Run("Create valid SEO record succeeds and normalizes slug", func(t *testing.T) {
		req := &CreateSEORequest{
			Slug:        "/About-Us/",
			Title:       "About Us Page",
			Description: "Learn more about our company",
			Keywords:    "about, company",
		}
		res, err := svc.Create(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Slug != "about-us" {
			t.Fatalf("expected slug 'about-us', got %s", res.Slug)
		}
		if res.OGType != "website" {
			t.Fatalf("expected default og_type 'website', got %s", res.OGType)
		}
	})

	t.Run("Get SEO by Slug succeeds", func(t *testing.T) {
		res, err := svc.GetBySlug(ctx, "about-us")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Title != "About Us Page" {
			t.Fatalf("expected title 'About Us Page', got %s", res.Title)
		}
	})

	t.Run("Get SEO by Slug not found returns ErrNotFound", func(t *testing.T) {
		_, err := svc.GetBySlug(ctx, "non-existent")
		if err != common.ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Update SEO record succeeds", func(t *testing.T) {
		newTitle := "Updated About Us Title"
		req := &UpdateSEORequest{
			Title: &newTitle,
		}
		updated, err := svc.Update(ctx, "seo-1", req)
		if err != nil {
			t.Fatalf("unexpected error on update: %v", err)
		}
		if updated.Title != newTitle {
			t.Fatalf("expected title %s, got %s", newTitle, updated.Title)
		}
	})

	t.Run("Delete SEO record succeeds", func(t *testing.T) {
		err := svc.Delete(ctx, "seo-1")
		if err != nil {
			t.Fatalf("unexpected error on delete: %v", err)
		}
		_, err = svc.GetByID(ctx, "seo-1")
		if err != common.ErrNotFound {
			t.Fatalf("expected ErrNotFound after deletion, got %v", err)
		}
	})
}
