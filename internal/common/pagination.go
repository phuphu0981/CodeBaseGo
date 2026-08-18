package common

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Pagination default limits and bounds.
const (
	DefaultPage    = 1
	DefaultPerPage = 10
	MaxPerPage     = 100
)

type PaginationQuery struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPaginationQuery(c *gin.Context) PaginationQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return PaginationQuery{Page: page, PerPage: perPage}
}

func (p PaginationQuery) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func NewPaginationMeta(query PaginationQuery, total int) PaginationMeta {
	return PaginationMeta{
		Page:       query.Page,
		PerPage:    query.PerPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(query.PerPage))),
	}
}

// CursorQuery contains parameters for cursor-based pagination.
type CursorQuery struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

// CursorMeta contains response metadata for cursor-based pagination.
type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

func NewCursorQuery(c *gin.Context) CursorQuery {
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("per_page", strconv.Itoa(DefaultPerPage))))
	if limit < 1 {
		limit = DefaultPerPage
	}
	if limit > MaxPerPage {
		limit = MaxPerPage
	}
	return CursorQuery{Cursor: cursor, Limit: limit}
}

func EncodeCursor(createdAt time.Time, id string) string {
	raw := fmt.Sprintf("%s|%s", createdAt.Format(time.RFC3339Nano), id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(cursorStr string) (time.Time, string, error) {
	bytes, err := base64.StdEncoding.DecodeString(cursorStr)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor encoding: %w", err)
	}

	parts := strings.SplitN(string(bytes), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format")
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor time format: %w", err)
	}

	return t, parts[1], nil
}
