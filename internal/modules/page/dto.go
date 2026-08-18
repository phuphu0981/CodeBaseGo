package page

import (
	"strings"

	"codebasego/internal/common"
)

// CreatePageRequest defines the payload for creating a new page.
type CreatePageRequest struct {
	Slug     string `json:"slug" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Status   string `json:"status"`
	Template string `json:"template"`
	Content  string `json:"content"`
}

func (r *CreatePageRequest) Validate() error {
	if strings.TrimSpace(r.Slug) == "" {
		return common.NewAppError(400, "slug cannot be empty")
	}
	if strings.TrimSpace(r.Title) == "" {
		return common.NewAppError(400, "title cannot be empty")
	}
	return nil
}

// UpdatePageRequest defines the payload for updating an existing page.
type UpdatePageRequest struct {
	Slug     *string `json:"slug" binding:"omitempty"`
	Title    *string `json:"title" binding:"omitempty"`
	Status   *string `json:"status"`
	Template *string `json:"template"`
	Content  *string `json:"content"`
}

func (r *UpdatePageRequest) Validate() error {
	if r.Slug != nil && strings.TrimSpace(*r.Slug) == "" {
		return common.NewAppError(400, "slug cannot be empty")
	}
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return common.NewAppError(400, "title cannot be empty")
	}
	return nil
}

// PageResponse represents the public representation of a page.
type PageResponse struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Template  string `json:"template"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
