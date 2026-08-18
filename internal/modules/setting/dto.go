package setting

import (
	"strings"

	"codebasego/internal/common"
)

// SetConfigRequest represents the payload to create or update a configuration path.
type SetConfigRequest struct {
	Scope   string `json:"scope"`
	ScopeID string `json:"scope_id"`
	Path    string `json:"path" binding:"required"`
	Value   string `json:"value"`
}

func (r *SetConfigRequest) Validate() error {
	if strings.TrimSpace(r.Path) == "" {
		return common.NewAppError(400, "config path cannot be empty")
	}
	return nil
}

// ConfigResponse represents the public response format for a configuration record.
type ConfigResponse struct {
	ID        string `json:"id"`
	Scope     string `json:"scope"`
	ScopeID   string `json:"scope_id"`
	Path      string `json:"path"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PublicConfigResponse provides safe, non-sensitive configuration values for Frontend consumption.
type PublicConfigResponse struct {
	BaseURL     string            `json:"base_url"`
	APIBaseURL  string            `json:"api_base_url"`
	StoreName   string            `json:"store_name"`
	Timezone    string            `json:"timezone"`
	CustomPaths map[string]string `json:"custom_paths,omitempty"`
}
