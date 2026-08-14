package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"codebasego/internal/common"
	"codebasego/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// Register creates a new user account and returns Access & Refresh tokens.
//
//	@Summary	Register a new user
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		RegisterRequest	true	"Registration info"
//	@Success	201		{object}	response.Body{data=TokenResponse}
//	@Failure	400		{object}	response.Body
//	@Failure	409		{object}	response.Body
//	@Router		/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	entity, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, common.ErrConflict) {
			response.Error(c, http.StatusConflict, "email already exists")
			return
		}
		response.InternalServerError(c)
		return
	}

	tokenPair, err := h.service.GenerateTokenPair(c.Request.Context(), entity.ID, entity.Email)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Created(c, tokenPair)
}

// Login authenticates a user and returns Access & Refresh tokens.
//
//	@Summary	Login
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		LoginRequest	true	"Login credentials"
//	@Success	200		{object}	response.Body{data=TokenResponse}
//	@Failure	401		{object}	response.Body
//	@Router		/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	entity, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokenPair, err := h.service.GenerateTokenPair(c.Request.Context(), entity.ID, entity.Email)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Success(c, tokenPair)
}

// Refresh generates a new Access & Refresh token pair using a valid Refresh token.
//
//	@Summary	Refresh Access Token
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		RefreshTokenRequest	true	"Refresh token"
//	@Success	200		{object}	response.Body{data=TokenResponse}
//	@Failure	401		{object}	response.Body
//	@Router		/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	tokenPair, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	response.Success(c, tokenPair)
}

// Logout revokes the given Refresh token.
//
//	@Summary	Logout
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		LogoutRequest	true	"Refresh token to revoke"
//	@Success	200		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Router		/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, http.StatusBadRequest, "failed to revoke token")
		return
	}

	response.Success(c, gin.H{"message": "logged out successfully"})
}

// LogoutAll revokes all active Refresh tokens for the authenticated user.
//
//	@Summary	Logout from all devices
//	@Tags		auth
//	@Produce	json
//	@Success	200	{object}	response.Body
//	@Failure	401	{object}	response.Body
//	@Security	BearerAuth
//	@Router		/auth/logout-all [post]
func (h *Handler) LogoutAll(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.LogoutAll(c.Request.Context(), userID); err != nil {
		response.InternalServerError(c)
		return
	}

	response.Success(c, gin.H{"message": "logged out from all devices successfully"})
}
