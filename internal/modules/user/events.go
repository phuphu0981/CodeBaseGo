package user

import "time"

const (
	EventUserRegistered = "user.registered"
	EventUserCreated    = "user.created"
	EventUserUpdated    = "user.updated"
	EventUserDeleted    = "user.deleted"
)

// UserRegisteredPayload is the typed payload for EventUserRegistered domain event.
type UserRegisteredPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
