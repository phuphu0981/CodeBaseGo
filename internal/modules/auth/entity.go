package auth

import "time"

// RefreshToken stores active and revoked refresh tokens in DB.
type RefreshToken struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);index;index:idx_user_revoked;not null"`
	Token     string    `json:"-" gorm:"uniqueIndex;type:varchar(64);not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"index;not null"`
	Revoked   bool      `json:"revoked" gorm:"default:false;index:idx_user_revoked;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
