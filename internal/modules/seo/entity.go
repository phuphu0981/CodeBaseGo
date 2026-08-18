package seo

import "time"

// SEO represents the SEO metadata domain model and GORM schema for pages.
type SEO struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Slug           string    `json:"slug" gorm:"uniqueIndex;type:varchar(255);not null"`
	Title          string    `json:"title" gorm:"type:varchar(255);not null"`
	Description    string    `json:"description" gorm:"type:text"`
	Keywords       string    `json:"keywords" gorm:"type:varchar(500)"`
	CanonicalURL   string    `json:"canonical_url" gorm:"type:varchar(500)"`
	OGTitle        string    `json:"og_title" gorm:"type:varchar(255)"`
	OGDescription  string    `json:"og_description" gorm:"type:text"`
	OGImage        string    `json:"og_image" gorm:"type:varchar(500)"`
	OGType         string    `json:"og_type" gorm:"type:varchar(50);default:'website'"`
	TwitterCard    string    `json:"twitter_card" gorm:"type:varchar(50);default:'summary_large_image'"`
	Robots         string    `json:"robots" gorm:"type:varchar(100);default:'index, follow'"`
	StructuredData string    `json:"structured_data" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time `json:"updated_at"`
}
