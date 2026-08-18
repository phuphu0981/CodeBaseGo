package page

import "time"

// Page represents the dynamic CMS/landing page domain model and GORM schema.
type Page struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;type:varchar(255);not null"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'published';index"`
	Template  string    `json:"template" gorm:"type:varchar(50);default:'default'"`
	Content   string    `json:"content" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
}
