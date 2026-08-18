package seo

import (
	"strings"

	"codebasego/internal/common"
)

// CreateSEORequest defines the request payload for creating an SEO page record.
type CreateSEORequest struct {
	Slug           string `json:"slug" binding:"required"`
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description"`
	Keywords       string `json:"keywords"`
	CanonicalURL   string `json:"canonical_url"`
	OGTitle        string `json:"og_title"`
	OGDescription  string `json:"og_description"`
	OGImage        string `json:"og_image"`
	OGType         string `json:"og_type"`
	TwitterCard    string `json:"twitter_card"`
	Robots         string `json:"robots"`
	StructuredData string `json:"structured_data"`
}

func (r *CreateSEORequest) Validate() error {
	if strings.TrimSpace(r.Slug) == "" {
		return common.NewAppError(400, "slug cannot be empty")
	}
	if strings.TrimSpace(r.Title) == "" {
		return common.NewAppError(400, "title cannot be empty")
	}
	return nil
}

// UpdateSEORequest defines the request payload for updating an SEO page record.
type UpdateSEORequest struct {
	Slug           *string `json:"slug" binding:"omitempty"`
	Title          *string `json:"title" binding:"omitempty"`
	Description    *string `json:"description"`
	Keywords       *string `json:"keywords"`
	CanonicalURL   *string `json:"canonical_url"`
	OGTitle        *string `json:"og_title"`
	OGDescription  *string `json:"og_description"`
	OGImage        *string `json:"og_image"`
	OGType         *string `json:"og_type"`
	TwitterCard    *string `json:"twitter_card"`
	Robots         *string `json:"robots"`
	StructuredData *string `json:"structured_data"`
}

func (r *UpdateSEORequest) Validate() error {
	if r.Slug != nil && strings.TrimSpace(*r.Slug) == "" {
		return common.NewAppError(400, "slug cannot be empty")
	}
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return common.NewAppError(400, "title cannot be empty")
	}
	return nil
}

// SEOResponse represents the public SEO metadata returned by the API.
type SEOResponse struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Keywords       string `json:"keywords"`
	CanonicalURL   string `json:"canonical_url"`
	OGTitle        string `json:"og_title"`
	OGDescription  string `json:"og_description"`
	OGImage        string `json:"og_image"`
	OGType         string `json:"og_type"`
	TwitterCard    string `json:"twitter_card"`
	Robots         string `json:"robots"`
	StructuredData string `json:"structured_data"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
