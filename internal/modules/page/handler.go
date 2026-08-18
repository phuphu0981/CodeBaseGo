package page

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

// List returns all pages. Supports both cursor-based and offset pagination.
//
//	@Summary	List pages
//	@Tags		pages
//	@Produce	json
//	@Param		cursor		query		string	false	"Opaque cursor for pagination"
//	@Param		limit		query		int		false	"Items per page for cursor pagination"	default(10)
//	@Param		page		query		int		false	"Page number for offset pagination"		default(1)
//	@Param		per_page	query		int		false	"Items per page for offset pagination"	default(10)
//	@Success	200			{object}	response.Body
//	@Router		/pages [get]
func (h *Handler) List(c *gin.Context) {
	if c.Query("cursor") != "" || c.Query("page") == "" {
		cursorQuery := common.NewCursorQuery(c)
		pages, meta, err := h.service.ListCursor(c.Request.Context(), cursorQuery)
		if err != nil {
			response.InternalServerError(c)
			return
		}

		results := make([]PageResponse, len(pages))
		for i, p := range pages {
			results[i] = toResponse(&p)
		}
		response.SuccessWithMeta(c, results, meta)
		return
	}

	paginationQuery := common.NewPaginationQuery(c)
	pages, total, err := h.service.List(c.Request.Context(), paginationQuery)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	results := make([]PageResponse, len(pages))
	for i, p := range pages {
		results[i] = toResponse(&p)
	}
	meta := common.NewPaginationMeta(paginationQuery, total)
	response.SuccessWithMeta(c, results, meta)
}

// GetByID returns a single page by its ID.
//
//	@Summary	Get page by ID
//	@Tags		pages
//	@Produce	json
//	@Param		id	path		string	true	"Page ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Router		/pages/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(p))
}

// GetBySlug returns page content and layout by slug.
//
//	@Summary	Get page by slug
//	@Tags		pages
//	@Produce	json
//	@Param		slug	query		string	true	"Page slug (e.g. 'home', 'about-us')"
//	@Success	200		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Failure	404		{object}	response.Body
//	@Router		/pages/by-slug [get]
func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		slug = c.Param("slug")
	}
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "slug query parameter is required")
		return
	}

	p, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(p))
}

// GetSlugs returns all published page slugs (useful for Next.js generateStaticParams).
//
//	@Summary	Get list of published page slugs for SSG
//	@Tags		pages
//	@Produce	json
//	@Success	200	{object}	response.Body
//	@Router		/pages/slugs [get]
func (h *Handler) GetSlugs(c *gin.Context) {
	slugs, err := h.service.GetPublishedSlugs(c.Request.Context())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, slugs)
}

// Create creates a new page.
//
//	@Summary	Create page
//	@Tags		pages
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CreatePageRequest	true	"Page payload"
//	@Success	201		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Failure	409		{object}	response.Body
//	@Security	BearerAuth
//	@Router		/pages [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	p, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, toResponse(p))
}

// Update modifies an existing page.
//
//	@Summary	Update page
//	@Tags		pages
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"Page ID"
//	@Param		body	body		UpdatePageRequest	true	"Page update payload"
//	@Success	200		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Failure	404		{object}	response.Body
//	@Security	BearerAuth
//	@Router		/pages/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	p, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(p))
}

// Delete removes a page.
//
//	@Summary	Delete page
//	@Tags		pages
//	@Produce	json
//	@Param		id	path		string	true	"Page ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Security	BearerAuth
//	@Router		/pages/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "page deleted"})
}

func toResponse(e *Page) PageResponse {
	return PageResponse{
		ID:        e.ID,
		Slug:      e.Slug,
		Title:     e.Title,
		Status:    e.Status,
		Template:  e.Template,
		Content:   e.Content,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}
