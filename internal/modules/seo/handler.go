package seo

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

// List returns all SEO records. Supports both cursor-based pagination and offset pagination.
//
//	@Summary	List SEO records
//	@Tags		seo
//	@Produce	json
//	@Param		cursor		query		string	false	"Opaque cursor for pagination"
//	@Param		limit		query		int		false	"Items per page for cursor pagination"	default(10)
//	@Param		page		query		int		false	"Page number for offset pagination"		default(1)
//	@Param		per_page	query		int		false	"Items per page for offset pagination"	default(10)
//	@Success	200			{object}	response.Body
//	@Router		/seo [get]
func (h *Handler) List(c *gin.Context) {
	if c.Query("cursor") != "" || c.Query("page") == "" {
		cursorQuery := common.NewCursorQuery(c)
		records, meta, err := h.service.ListCursor(c.Request.Context(), cursorQuery)
		if err != nil {
			response.InternalServerError(c)
			return
		}

		results := make([]SEOResponse, len(records))
		for i, r := range records {
			results[i] = toResponse(&r)
		}
		response.SuccessWithMeta(c, results, meta)
		return
	}

	paginationQuery := common.NewPaginationQuery(c)
	records, total, err := h.service.List(c.Request.Context(), paginationQuery)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	results := make([]SEOResponse, len(records))
	for i, r := range records {
		results[i] = toResponse(&r)
	}
	meta := common.NewPaginationMeta(paginationQuery, total)
	response.SuccessWithMeta(c, results, meta)
}

// GetByID returns a single SEO record by its ID.
//
//	@Summary	Get SEO record by ID
//	@Tags		seo
//	@Produce	json
//	@Param		id	path		string	true	"SEO Record ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Router		/seo/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	record, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(record))
}

// GetBySlug returns an SEO metadata record by page slug.
//
//	@Summary	Get SEO record by slug
//	@Tags		seo
//	@Produce	json
//	@Param		slug	query		string	true	"Page slug (e.g. 'home', 'about-us', 'products/shoes')"
//	@Success	200		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Failure	404		{object}	response.Body
//	@Router		/seo/by-slug [get]
func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		slug = c.Param("slug")
	}
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "slug query parameter is required")
		return
	}

	record, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(record))
}

// Create creates a new SEO record for a page.
//
//	@Summary	Create SEO record
//	@Tags		seo
//	@Accept		json
//	@Produce	json
//	@Param		body	body		CreateSEORequest	true	"SEO Metadata payload"
//	@Success	201		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Failure	409		{object}	response.Body
//	@Security	BearerAuth
//	@Router		/seo [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateSEORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	record, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, toResponse(record))
}

// Update modifies an existing SEO record.
//
//	@Summary	Update SEO record
//	@Tags		seo
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"SEO Record ID"
//	@Param		body	body		UpdateSEORequest	true	"SEO Update payload"
//	@Success	200		{object}	response.Body
//	@Failure	400		{object}	response.Body
//	@Failure	404		{object}	response.Body
//	@Security	BearerAuth
//	@Router		/seo/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateSEORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	record, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, toResponse(record))
}

// Delete removes an SEO record.
//
//	@Summary	Delete SEO record
//	@Tags		seo
//	@Produce	json
//	@Param		id	path		string	true	"SEO Record ID"
//	@Success	200	{object}	response.Body
//	@Failure	404	{object}	response.Body
//	@Security	BearerAuth
//	@Router		/seo/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "seo record deleted"})
}

func toResponse(e *SEO) SEOResponse {
	return SEOResponse{
		ID:             e.ID,
		Slug:           e.Slug,
		Title:          e.Title,
		Description:    e.Description,
		Keywords:       e.Keywords,
		CanonicalURL:   e.CanonicalURL,
		OGTitle:        e.OGTitle,
		OGDescription:  e.OGDescription,
		OGImage:        e.OGImage,
		OGType:         e.OGType,
		TwitterCard:    e.TwitterCard,
		Robots:         e.Robots,
		StructuredData: e.StructuredData,
		CreatedAt:      e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      e.UpdatedAt.Format(time.RFC3339),
	}
}
