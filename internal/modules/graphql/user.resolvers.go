package graphql

import (
	"context"
	"time"

	"codebasego/internal/common"
	"codebasego/internal/modules/user"
)

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	currentUserID, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if currentUserID != id {
		return nil, common.ErrForbidden
	}

	req := &user.UpdateRequest{
		Email:    input.Email,
		Password: input.Password,
		Name:     input.Name,
	}

	u, err := r.userService.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	return toGQLUser(u), nil
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (*Message, error) {
	currentUserID, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if currentUserID != id {
		return nil, common.ErrForbidden
	}

	if err := r.userService.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &Message{Message: "user deleted"}, nil
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*User, error) {
	currentUserID, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}

	u, err := r.userService.GetByID(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	return toGQLUser(u), nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, page *int, limit *int) ([]*User, error) {
	if _, err := RequireAuth(ctx); err != nil {
		return nil, err
	}

	p := 1
	if page != nil && *page > 0 {
		p = *page
	}
	l := 10
	if limit != nil && *limit > 0 {
		l = *limit
	}
	if l > 100 {
		l = 100
	}

	users, _, err := r.userService.List(ctx, common.PaginationQuery{Page: p, PerPage: l})
	if err != nil {
		return nil, err
	}

	gqlUsers := make([]*User, len(users))
	for i, u := range users {
		gqlUsers[i] = toGQLUser(&u)
	}
	return gqlUsers, nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id string) (*User, error) {
	if _, err := RequireAuth(ctx); err != nil {
		return nil, err
	}

	u, err := r.userService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toGQLUser(u), nil
}

func toGQLUser(e *user.User) *User {
	return &User{
		ID:        e.ID,
		Email:     e.Email,
		Name:      e.Name,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}
