package graphql

import (
	"context"
)

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, input RegisterInput) (*TokenPair, error) {
	u, err := r.authService.Register(ctx, input.Email, input.Password, input.Name)
	if err != nil {
		return nil, err
	}

	pair, err := r.authService.GenerateTokenPair(ctx, u.ID, u.Email, u.Role)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	}, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	u, err := r.authService.Login(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	pair, err := r.authService.GenerateTokenPair(ctx, u.ID, u.Email, u.Role)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	}, nil
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	pair, err := r.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	}, nil
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context, refreshToken string) (*Message, error) {
	if err := r.authService.Logout(ctx, refreshToken); err != nil {
		return nil, err
	}
	return &Message{Message: "logged out successfully"}, nil
}

// LogoutAll is the resolver for the logoutAll field.
func (r *mutationResolver) LogoutAll(ctx context.Context) (*Message, error) {
	currentUserID, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.authService.LogoutAll(ctx, currentUserID); err != nil {
		return nil, err
	}
	return &Message{Message: "logged out from all devices successfully"}, nil
}
