package seo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"codebasego/internal/common"
	"codebasego/internal/platform/database"
)

// GormRepository implements Repository interface using GORM.
type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll(ctx context.Context, query common.PaginationQuery) ([]SEO, int, error) {
	var records []SEO
	var total int64

	db := database.GetDB(ctx, r.db)
	if err := db.Model(&SEO{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Offset(query.Offset()).Limit(query.PerPage).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, int(total), nil
}

func (r *GormRepository) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]SEO, common.CursorMeta, error) {
	var records []SEO
	db := database.GetDB(ctx, r.db)

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}

	tx := db.Model(&SEO{})
	if query.Cursor != "" {
		t, lastID, err := common.DecodeCursor(query.Cursor)
		if err == nil {
			tx = tx.Where("created_at < ? OR (created_at = ? AND id < ?)", t, t, lastID)
		}
	}

	if err := tx.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&records).Error; err != nil {
		return nil, common.CursorMeta{}, err
	}

	hasMore := false
	var nextCursor string
	if len(records) > limit {
		hasMore = true
		records = records[:limit]
	}

	if len(records) > 0 && hasMore {
		lastRecord := records[len(records)-1]
		nextCursor = common.EncodeCursor(lastRecord.CreatedAt, lastRecord.ID)
	}

	meta := common.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}

	return records, meta, nil
}

func (r *GormRepository) FindByID(ctx context.Context, id string) (*SEO, error) {
	var s SEO
	if err := database.GetDB(ctx, r.db).First(&s, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *GormRepository) FindBySlug(ctx context.Context, slug string) (*SEO, error) {
	var s SEO
	if err := database.GetDB(ctx, r.db).First(&s, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *GormRepository) Create(ctx context.Context, entity *SEO) error {
	if entity.ID == "" {
		entity.ID = uuid.New().String()
	}
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	if err := database.GetDB(ctx, r.db).Create(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isDuplicateError(err) {
			return common.ErrConflict
		}
		return err
	}
	return nil
}

func (r *GormRepository) Update(ctx context.Context, entity *SEO) error {
	entity.UpdatedAt = time.Now()
	updates := map[string]interface{}{
		"slug":            entity.Slug,
		"title":           entity.Title,
		"description":     entity.Description,
		"keywords":        entity.Keywords,
		"canonical_url":   entity.CanonicalURL,
		"og_title":        entity.OGTitle,
		"og_description":  entity.OGDescription,
		"og_image":        entity.OGImage,
		"og_type":         entity.OGType,
		"twitter_card":    entity.TwitterCard,
		"robots":          entity.Robots,
		"structured_data": entity.StructuredData,
		"updated_at":      entity.UpdatedAt,
	}
	res := database.GetDB(ctx, r.db).Model(&SEO{}).Where("id = ?", entity.ID).Updates(updates)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrDuplicatedKey) || isDuplicateError(res.Error) {
			return common.ErrConflict
		}
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := database.GetDB(ctx, r.db).Model(&SEO{}).Where("id = ?", entity.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return common.ErrNotFound
		}
	}
	return nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	res := database.GetDB(ctx, r.db).Delete(&SEO{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "1062") ||
		strings.Contains(msg, "23505")
}
