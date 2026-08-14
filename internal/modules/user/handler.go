package user

import (
	"net/http"
	"time"

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

// List returns all users. Supports both cursor-based pagination (default/high-performance) and offset pagination.
//
//	@Summary	List users
//	@Tags		users
//	@Produce	json
//	@Param		cursor		query		string	false	"Opaque cursor for pagination"
//	@Param		limit		query		int		false	"Items per page for cursor pagination"	default(10)
//	@Param		page		query		int		false	"Page number for offset pagination"		default(1)
//	@Param		per_page	query		int		false	"Items per page for offset pagination"	default(10)
//	@Success	200			{object}	response.Body
//	@Security	BearerAuth
//	@Router		/users [get]
func (h *Handler) List(c *gin.Context) {
	if c.Query("cursor") != "" || c.Query("page") == "" {
		cursorQuery := common.NewCursorQuery(c)
		users, meta, err := h.service.ListCursor(c.Request.Context(), cursorQuery)
		if err != nil {
			response.InternalServerError(c)
			return
		}

		results := make([]Response, len(users))
		for i, u := range users {
			results[i] = toResponse(&u)
		}
		response.SuccessWithMeta(c, results, meta)
		return
	}

	paginationQuery := common.NewPaginationQuery(c)
	users, total, err := h.service.List(c.Request.Context(), paginationQuery)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	results := make([]Response, len(users))
	for i, u := range users {
		results[i] = toResponse(&u)
	}
	meta := common.NewPaginationMeta(paginationQuery, total)
	response.SuccessWithMeta(c, results, meta)
}

// GetByID returns a single user.
//
//	@Summary	Get user by ID
//	@Tags		users
//	@Produce	json
//	@Param		id	path		string	true	"User ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Security	BearerAuth
//	@Router		/users/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	u, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(u))
}

// Update modifies an existing user.
//
//	@Summary	Update user
//	@Tags		users
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string			true	"User ID"
//	@Param		body	body		UpdateRequest	true	"Update fields"
//	@Success	200		{object}	response.Body
//	@Failure	404		{object}	response.Body
//	@Security	BearerAuth
//	@Router		/users/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	targetID := c.Param("id")
	if currentUserID := c.GetString("user_id"); currentUserID != targetID {
		response.Error(c, http.StatusForbidden, "forbidden: cannot modify another user's account")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	u, err := h.service.Update(c.Request.Context(), targetID, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(u))
}

// Delete removes a user.
//
//	@Summary	Delete user
//	@Tags		users
//	@Produce	json
//	@Param		id	path		string	true	"User ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Security	BearerAuth
//	@Router		/users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	targetID := c.Param("id")
	if currentUserID := c.GetString("user_id"); currentUserID != targetID {
		response.Error(c, http.StatusForbidden, "forbidden: cannot modify another user's account")
		return
	}

	if err := h.service.Delete(c.Request.Context(), targetID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "user deleted"})
}

func toResponse(e *User) Response {
	return Response{
		ID:        e.ID,
		Email:     e.Email,
		Name:      e.Name,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}
