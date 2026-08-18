package setting

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

// GetPublic returns non-sensitive public system configurations (Base URL, Store Name, Timezone).
//
//	@Summary	Get public system configs
//	@Tags		settings
//	@Produce	json
//	@Success	200	{object}	response.Body
//	@Router		/settings/public [get]
func (h *Handler) GetPublic(c *gin.Context) {
	configs, err := h.service.GetPublicConfigs(c.Request.Context())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, configs)
}

// GetByPath returns a single configuration record by its path and scope.
//
//	@Summary	Get config by path
//	@Tags		settings
//	@Produce	json
//	@Param		path		query		string	true	"Config Path (e.g. web/unsecure/base_url)"
//	@Param		scope		query		string	false	"Scope"		default(default)
//	@Param		scope_id	query		string	false	"Scope ID"	default(0)
//	@Success	200			{object}	response.Body
//	@Failure	400			{object}	response.Body
//	@Failure	404			{object}	response.Body
//	@Router		/settings/by-path [get]
func (h *Handler) GetByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, "path query parameter is required")
		return
	}

	scope := c.DefaultQuery("scope", "default")
	scopeID := c.DefaultQuery("scope_id", "0")

	record, err := h.service.GetByPath(c.Request.Context(), scope, scopeID, path)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(record))
}

// List returns all system configuration records.
//
//	@Summary	List system configurations
//	@Tags		settings
//	@Produce	json
//	@Param		path_prefix	query		string	false	"Filter by path prefix (e.g. web/)"
//	@Param		scope		query		string	false	"Filter by scope"
//	@Param		cursor		query		string	false	"Opaque cursor for pagination"
//	@Param		limit		query		int		false	"Items per page for cursor pagination"	default(10)
//	@Param		page		query		int		false	"Page number for offset pagination"		default(1)
//	@Param		per_page	query		int		false	"Items per page for offset pagination"	default(10)
//	@Success	200			{object}	response.Body
//	@Security	BearerAuth
//	@Router		/settings [get]
func (h *Handler) List(c *gin.Context) {
	scope := c.Query("scope")
	pathPrefix := c.Query("path_prefix")

	if c.Query("cursor") != "" || c.Query("page") == "" {
		cursorQuery := common.NewCursorQuery(c)
		records, meta, err := h.service.ListCursor(c.Request.Context(), cursorQuery, scope, pathPrefix)
		if err != nil {
			response.InternalServerError(c)
			return
		}

		results := make([]ConfigResponse, len(records))
		for i, r := range records {
			results[i] = toResponse(&r)
		}
		response.SuccessWithMeta(c, results, meta)
		return
	}

	paginationQuery := common.NewPaginationQuery(c)
	records, total, err := h.service.List(c.Request.Context(), paginationQuery, scope, pathPrefix)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	results := make([]ConfigResponse, len(records))
	for i, r := range records {
		results[i] = toResponse(&r)
	}
	meta := common.NewPaginationMeta(paginationQuery, total)
	response.SuccessWithMeta(c, results, meta)
}

// GetByID returns a configuration record by ID.
//
//	@Summary	Get config by ID
//	@Tags		settings
//	@Produce	json
//	@Param		id	path		string	true	"Config ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Security	BearerAuth
//	@Router		/settings/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	record, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(record))
}

// Set creates or updates a configuration value for a path.
//
//	@Summary	Set configuration value
//	@Tags		settings
//	@Accept		json
//	@Produce	json
//	@Param		body	body		SetConfigRequest	true	"Config payload"
//	@Success	200		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Security	BearerAuth
//	@Router		/settings [post]
func (h *Handler) Set(c *gin.Context) {
	var req SetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	record, err := h.service.Set(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(record))
}

// Delete removes a configuration record by ID.
//
//	@Summary	Delete config by ID
//	@Tags		settings
//	@Produce	json
//	@Param		id	path		string	true	"Config ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Security	BearerAuth
//	@Router		/settings/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "config deleted"})
}

func toResponse(e *CoreConfig) ConfigResponse {
	return ConfigResponse{
		ID:        e.ID,
		Scope:     e.Scope,
		ScopeID:   e.ScopeID,
		Path:      e.Path,
		Value:     e.Value,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}
