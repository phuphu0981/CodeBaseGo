package setting

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"codebasego/internal/common"
	"codebasego/internal/platform/database"
)

// GormRepository implements Repository interface for core_config_data using GORM.
type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll(ctx context.Context, query common.PaginationQuery, scope, pathPrefix string) ([]CoreConfig, int, error) {
	var records []CoreConfig
	var total int64

	db := database.GetDB(ctx, r.db).Model(&CoreConfig{})
	if scope != "" {
		db = db.Where("scope = ?", scope)
	}
	if pathPrefix != "" {
		db = db.Where("path LIKE ?", pathPrefix+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("path ASC").Offset(query.Offset()).Limit(query.PerPage).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, int(total), nil
}

func (r *GormRepository) FindAllCursor(ctx context.Context, query common.CursorQuery, scope, pathPrefix string) ([]CoreConfig, common.CursorMeta, error) {
	var records []CoreConfig
	db := database.GetDB(ctx, r.db).Model(&CoreConfig{})

	if scope != "" {
		db = db.Where("scope = ?", scope)
	}
	if pathPrefix != "" {
		db = db.Where("path LIKE ?", pathPrefix+"%")
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}

	tx := db
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

func (r *GormRepository) FindByID(ctx context.Context, id string) (*CoreConfig, error) {
	var c CoreConfig
	if err := database.GetDB(ctx, r.db).First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) FindByPath(ctx context.Context, scope, scopeID, path string) (*CoreConfig, error) {
	var c CoreConfig
	if scope == "" {
		scope = "default"
	}
	if scopeID == "" {
		scopeID = "0"
	}

	if err := database.GetDB(ctx, r.db).First(&c, "scope = ? AND scope_id = ? AND path = ?", scope, scopeID, path).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) FindByPrefix(ctx context.Context, scope, scopeID, prefix string) ([]CoreConfig, error) {
	var records []CoreConfig
	if scope == "" {
		scope = "default"
	}
	if scopeID == "" {
		scopeID = "0"
	}

	if err := database.GetDB(ctx, r.db).Where("scope = ? AND scope_id = ? AND path LIKE ?", scope, scopeID, prefix+"%").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *GormRepository) Save(ctx context.Context, entity *CoreConfig) error {
	if entity.Scope == "" {
		entity.Scope = "default"
	}
	if entity.ScopeID == "" {
		entity.ScopeID = "0"
	}

	now := time.Now()
	var existing CoreConfig
	err := database.GetDB(ctx, r.db).First(&existing, "scope = ? AND scope_id = ? AND path = ?", entity.Scope, entity.ScopeID, entity.Path).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if entity.ID == "" {
				entity.ID = uuid.New().String()
			}
			entity.CreatedAt = now
			entity.UpdatedAt = now
			return database.GetDB(ctx, r.db).Create(entity).Error
		}
		return err
	}

	entity.ID = existing.ID
	entity.CreatedAt = existing.CreatedAt
	entity.UpdatedAt = now

	return database.GetDB(ctx, r.db).Model(&existing).Updates(map[string]interface{}{
		"value":      entity.Value,
		"updated_at": now,
	}).Error
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	res := database.GetDB(ctx, r.db).Delete(&CoreConfig{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}

func (r *GormRepository) DeleteByPath(ctx context.Context, scope, scopeID, path string) error {
	if scope == "" {
		scope = "default"
	}
	if scopeID == "" {
		scopeID = "0"
	}
	res := database.GetDB(ctx, r.db).Delete(&CoreConfig{}, "scope = ? AND scope_id = ? AND path = ?", scope, scopeID, path)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrNotFound
	}
	return nil
}
