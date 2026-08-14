package user

import "time"

// User represents the user domain model and GORM schema.
type User struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Email     string    `json:"email" gorm:"uniqueIndex;type:varchar(255);not null"`
	Password  string    `json:"-" gorm:"not null"`
	Name      string    `json:"name" gorm:"type:varchar(255)"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
}
