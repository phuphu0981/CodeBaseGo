package user

import (
	"context"
	"testing"

	"codebasego/internal/common"
)

type mockRepo struct{}

func (m *mockRepo) FindAll(ctx context.Context, query common.PaginationQuery) ([]User, int, error) {
	return nil, 0, nil
}
func (m *mockRepo) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]User, common.CursorMeta, error) {
	return nil, common.CursorMeta{}, nil
}
func (m *mockRepo) FindByID(ctx context.Context, id string) (*User, error) {
	return &User{ID: id, Email: "test@example.com", Name: "Test"}, nil
}
func (m *mockRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	return nil, common.ErrNotFound
}
func (m *mockRepo) Create(ctx context.Context, entity *User) error { return nil }
func (m *mockRepo) Update(ctx context.Context, entity *User) error { return nil }
func (m *mockRepo) Delete(ctx context.Context, id string) error     { return nil }

func TestUserServiceValidation(t *testing.T) {
	svc := NewService(&mockRepo{})
	ctx := context.Background()

	t.Run("Create user with invalid email fails", func(t *testing.T) {
		req := &CreateRequest{
			Email:    "invalid-email",
			Password: "password123",
			Name:     "Alice",
		}
		_, err := svc.Create(ctx, req, "hashed")
		if err == nil {
			t.Fatal("expected validation error for invalid email")
		}
	})

	t.Run("Create user with short password fails", func(t *testing.T) {
		req := &CreateRequest{
			Email:    "alice@example.com",
			Password: "123",
			Name:     "Alice",
		}
		_, err := svc.Create(ctx, req, "hashed")
		if err == nil {
			t.Fatal("expected validation error for short password")
		}
	})

	t.Run("Create user with empty name fails", func(t *testing.T) {
		req := &CreateRequest{
			Email:    "alice@example.com",
			Password: "password123",
			Name:     "",
		}
		_, err := svc.Create(ctx, req, "hashed")
		if err == nil {
			t.Fatal("expected validation error for empty name")
		}
	})

	t.Run("Create valid user succeeds", func(t *testing.T) {
		req := &CreateRequest{
			Email:    "alice@example.com",
			Password: "password123",
			Name:     "Alice",
		}
		u, err := svc.Create(ctx, req, "hashed")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Email != "alice@example.com" {
			t.Fatalf("expected email alice@example.com, got %s", u.Email)
		}
	})

	t.Run("Update user with invalid email fails", func(t *testing.T) {
		badEmail := "bademail"
		req := &UpdateRequest{Email: &badEmail}
		_, err := svc.Update(ctx, "1", req)
		if err == nil {
			t.Fatal("expected validation error on update")
		}
	})

	t.Run("Update user with short password fails", func(t *testing.T) {
		shortPass := "123"
		req := &UpdateRequest{Password: &shortPass}
		_, err := svc.Update(ctx, "1", req)
		if err == nil {
			t.Fatal("expected validation error on short password update")
		}
	})

	t.Run("Update user with valid password succeeds", func(t *testing.T) {
		newPass := "newpassword123"
		req := &UpdateRequest{Password: &newPass}
		u, err := svc.Update(ctx, "1", req)
		if err != nil {
			t.Fatalf("unexpected error on password update: %v", err)
		}
		if u.Password == "" || u.Password == newPass {
			t.Fatal("expected password to be hashed")
		}
	})
}
