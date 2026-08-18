package setting

import (
	"context"
	"strings"

	"codebasego/internal/common"
)

// Standard system configuration path constants.
const (
	PathBaseURL         = "web/unsecure/base_url"
	PathSecureBaseURL   = "web/secure/base_url"
	PathAPIBaseURL      = "web/api_base_url"
	PathStoreName       = "general/store_information/name"
	PathTimezone        = "general/locale/timezone"
	PathSEODefaultTitle = "seo/general/default_title"
	PathSEOTitleSuffix  = "seo/general/title_suffix"

	PrefixPublicPath = "web/public/"

	DefaultScope      = "default"
	DefaultScopeID    = "0"
	DefaultBaseURL    = "http://localhost:3000"
	DefaultAPIBaseURL = "http://localhost:8080/api/v1"
	DefaultStoreName  = "CodebaseGo Store"
	DefaultTimezone   = "Asia/Ho_Chi_Minh"
)

// Service contains system configuration business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, path string) (string, error) {
	record, err := s.repo.FindByPath(ctx, DefaultScope, DefaultScopeID, strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	return record.Value, nil
}

func (s *Service) GetWithDefault(ctx context.Context, path string, defaultValue string) string {
	val, err := s.Get(ctx, path)
	if err != nil || strings.TrimSpace(val) == "" {
		return defaultValue
	}
	return val
}

func (s *Service) GetBaseURL(ctx context.Context) string {
	return s.GetWithDefault(ctx, PathBaseURL, DefaultBaseURL)
}

func (s *Service) GetPublicConfigs(ctx context.Context) (*PublicConfigResponse, error) {
	baseURL := s.GetWithDefault(ctx, PathBaseURL, DefaultBaseURL)
	apiBaseURL := s.GetWithDefault(ctx, PathAPIBaseURL, DefaultAPIBaseURL)
	storeName := s.GetWithDefault(ctx, PathStoreName, DefaultStoreName)
	timezone := s.GetWithDefault(ctx, PathTimezone, DefaultTimezone)

	publicList, err := s.repo.FindByPrefix(ctx, DefaultScope, DefaultScopeID, PrefixPublicPath)
	customMap := make(map[string]string)
	if err == nil {
		for _, item := range publicList {
			customMap[item.Path] = item.Value
		}
	}

	return &PublicConfigResponse{
		BaseURL:     baseURL,
		APIBaseURL:  apiBaseURL,
		StoreName:   storeName,
		Timezone:    timezone,
		CustomPaths: customMap,
	}, nil
}

func (s *Service) List(ctx context.Context, query common.PaginationQuery, scope, prefix string) ([]CoreConfig, int, error) {
	return s.repo.FindAll(ctx, query, scope, prefix)
}

func (s *Service) ListCursor(ctx context.Context, query common.CursorQuery, scope, prefix string) ([]CoreConfig, common.CursorMeta, error) {
	return s.repo.FindAllCursor(ctx, query, scope, prefix)
}

func (s *Service) GetByID(ctx context.Context, id string) (*CoreConfig, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetByPath(ctx context.Context, scope, scopeID, path string) (*CoreConfig, error) {
	if scope == "" {
		scope = DefaultScope
	}
	if scopeID == "" {
		scopeID = DefaultScopeID
	}
	return s.repo.FindByPath(ctx, scope, scopeID, strings.TrimSpace(path))
}

func (s *Service) Set(ctx context.Context, req *SetConfigRequest) (*CoreConfig, error) {
	if req == nil {
		return nil, common.ErrBadRequest
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	scope := req.Scope
	if scope == "" {
		scope = DefaultScope
	}
	scopeID := req.ScopeID
	if scopeID == "" {
		scopeID = DefaultScopeID
	}

	entity := &CoreConfig{
		Scope:   scope,
		ScopeID: scopeID,
		Path:    strings.TrimSpace(req.Path),
		Value:   req.Value,
	}

	if err := s.repo.Save(ctx, entity); err != nil {
		return nil, err
	}

	return s.repo.FindByPath(ctx, entity.Scope, entity.ScopeID, entity.Path)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
