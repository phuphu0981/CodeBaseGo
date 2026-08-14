package user

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

func (r *GormRepository) FindAll(ctx context.Context, query common.PaginationQuery) ([]User, int, error) {
	var users []User
	var total int64

	db := database.GetDB(ctx, r.db)
	if err := db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Offset(query.Offset()).Limit(query.PerPage).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

func (r *GormRepository) FindAllCursor(ctx context.Context, query common.CursorQuery) ([]User, common.CursorMeta, error) {
	var users []User
	db := database.GetDB(ctx, r.db)

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}

	tx := db.Model(&User{})
	if query.Cursor != "" {
		t, lastID, err := common.DecodeCursor(query.Cursor)
		if err == nil {
			tx = tx.Where("created_at < ? OR (created_at = ? AND id < ?)", t, t, lastID)
		}
	}

	if err := tx.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&users).Error; err != nil {
		return nil, common.CursorMeta{}, err
	}

	hasMore := false
	var nextCursor string
	if len(users) > limit {
		hasMore = true
		users = users[:limit]
	}

	if len(users) > 0 && hasMore {
		lastUser := users[len(users)-1]
		nextCursor = common.EncodeCursor(lastUser.CreatedAt, lastUser.ID)
	}

	meta := common.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}

	return users, meta, nil
}

func (r *GormRepository) FindByID(ctx context.Context, id string) (*User, error) {
	var u User
	if err := database.GetDB(ctx, r.db).First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := database.GetDB(ctx, r.db).First(&u, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) Create(ctx context.Context, entity *User) error {
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

func (r *GormRepository) Update(ctx context.Context, entity *User) error {
	entity.UpdatedAt = time.Now()
	updates := map[string]interface{}{
		"email":      entity.Email,
		"name":       entity.Name,
		"updated_at": entity.UpdatedAt,
	}
	if entity.Password != "" {
		updates["password"] = entity.Password
	}
	res := database.GetDB(ctx, r.db).Model(&User{}).Where("id = ?", entity.ID).Updates(updates)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrDuplicatedKey) || isDuplicateError(res.Error) {
			return common.ErrConflict
		}
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := database.GetDB(ctx, r.db).Model(&User{}).Where("id = ?", entity.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return common.ErrNotFound
		}
	}
	return nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	res := database.GetDB(ctx, r.db).Delete(&User{}, "id = ?", id)
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

