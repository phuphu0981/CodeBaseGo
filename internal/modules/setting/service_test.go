package setting

import (
	"context"
	"testing"

	"codebasego/internal/common"
)

type mockSettingRepo struct {
	configs map[string]*CoreConfig
}

func newMockSettingRepo() *mockSettingRepo {
	return &mockSettingRepo{
		configs: make(map[string]*CoreConfig),
	}
}

func (m *mockSettingRepo) FindAll(ctx context.Context, query common.PaginationQuery, scope, pathPrefix string) ([]CoreConfig, int, error) {
	var list []CoreConfig
	for _, v := range m.configs {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockSettingRepo) FindAllCursor(ctx context.Context, query common.CursorQuery, scope, pathPrefix string) ([]CoreConfig, common.CursorMeta, error) {
	var list []CoreConfig
	for _, v := range m.configs {
		list = append(list, *v)
	}
	return list, common.CursorMeta{HasMore: false, Limit: query.Limit}, nil
}

func (m *mockSettingRepo) FindByID(ctx context.Context, id string) (*CoreConfig, error) {
	for _, v := range m.configs {
		if v.ID == id {
			return v, nil
		}
	}
	return nil, common.ErrNotFound
}

func (m *mockSettingRepo) FindByPath(ctx context.Context, scope, scopeID, path string) (*CoreConfig, error) {
	key := scope + "/" + scopeID + "/" + path
	if v, ok := m.configs[key]; ok {
		return v, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockSettingRepo) FindByPrefix(ctx context.Context, scope, scopeID, prefix string) ([]CoreConfig, error) {
	var list []CoreConfig
	for _, v := range m.configs {
		if v.Scope == scope && v.ScopeID == scopeID {
			list = append(list, *v)
		}
	}
	return list, nil
}

func (m *mockSettingRepo) Save(ctx context.Context, entity *CoreConfig) error {
	if entity.ID == "" {
		entity.ID = "cfg-1"
	}
	key := entity.Scope + "/" + entity.ScopeID + "/" + entity.Path
	m.configs[key] = entity
	return nil
}

func (m *mockSettingRepo) Delete(ctx context.Context, id string) error {
	for k, v := range m.configs {
		if v.ID == id {
			delete(m.configs, k)
			return nil
		}
	}
	return common.ErrNotFound
}

func (m *mockSettingRepo) DeleteByPath(ctx context.Context, scope, scopeID, path string) error {
	key := scope + "/" + scopeID + "/" + path
	if _, ok := m.configs[key]; ok {
		delete(m.configs, key)
		return nil
	}
	return common.ErrNotFound
}

func TestSettingService(t *testing.T) {
	repo := newMockSettingRepo()
	svc := NewService(repo)
	ctx := context.Background()

	t.Run("Set and Get Base URL config", func(t *testing.T) {
		req := &SetConfigRequest{
			Path:  PathBaseURL,
			Value: "https://mysite.com",
		}
		res, err := svc.Set(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error setting config: %v", err)
		}
		if res.Value != "https://mysite.com" {
			t.Fatalf("expected value 'https://mysite.com', got %s", res.Value)
		}

		val, err := svc.Get(ctx, PathBaseURL)
		if err != nil {
			t.Fatalf("unexpected error getting config: %v", err)
		}
		if val != "https://mysite.com" {
			t.Fatalf("expected %s, got %s", "https://mysite.com", val)
		}
	})

	t.Run("GetWithDefault returns fallback for non-existent path", func(t *testing.T) {
		fallback := svc.GetWithDefault(ctx, "non/existent/path", "default_val")
		if fallback != "default_val" {
			t.Fatalf("expected 'default_val', got %s", fallback)
		}
	})

	t.Run("GetBaseURL returns updated base URL", func(t *testing.T) {
		url := svc.GetBaseURL(ctx)
		if url != "https://mysite.com" {
			t.Fatalf("expected 'https://mysite.com', got %s", url)
		}
	})

	t.Run("GetPublicConfigs returns public structure", func(t *testing.T) {
		publicCfg, err := svc.GetPublicConfigs(ctx)
		if err != nil {
			t.Fatalf("unexpected error getting public configs: %v", err)
		}
		if publicCfg.BaseURL != "https://mysite.com" {
			t.Fatalf("expected base url 'https://mysite.com', got %s", publicCfg.BaseURL)
		}
	})

	t.Run("Set with empty path fails", func(t *testing.T) {
		req := &SetConfigRequest{Path: "  "}
		_, err := svc.Set(ctx, req)
		if err == nil {
			t.Fatal("expected error for empty path")
		}
	})
}
