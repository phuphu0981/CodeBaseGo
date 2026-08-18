package user

import "codebasego/internal/common"

// CreateRequest is the payload for creating a new user.
type CreateRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"omitempty"`
}

func (r *CreateRequest) Validate() error {
	if r.Email == "" || !common.IsValidEmail(r.Email) {
		return common.NewAppError(400, "invalid email format")
	}
	if len(r.Password) < 6 {
		return common.NewAppError(400, "password must be at least 6 characters long")
	}
	if r.Name == "" {
		return common.NewAppError(400, "name cannot be empty")
	}
	return nil
}

// UpdateRequest is the payload for updating a user. Nil fields are skipped.
type UpdateRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty,min=6"`
	Name     *string `json:"name" binding:"omitempty"`
	Role     *string `json:"role" binding:"omitempty"`
}

func (r *UpdateRequest) Validate() error {
	if r.Email != nil && (*r.Email == "" || !common.IsValidEmail(*r.Email)) {
		return common.NewAppError(400, "invalid email format")
	}
	if r.Password != nil && len(*r.Password) < 6 {
		return common.NewAppError(400, "password must be at least 6 characters long")
	}
	if r.Name != nil && *r.Name == "" {
		return common.NewAppError(400, "name cannot be empty")
	}
	return nil
}

// Response is the public user representation returned by the API.
type Response struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
