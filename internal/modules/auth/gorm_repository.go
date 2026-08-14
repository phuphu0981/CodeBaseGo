package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"codebasego/internal/common"
	"codebasego/internal/platform/database"
)

type GormRefreshTokenRepository struct {
	db *gorm.DB
}

func NewGormRefreshTokenRepository(db *gorm.DB) *GormRefreshTokenRepository {
	return &GormRefreshTokenRepository{db: db}
}

func (r *GormRefreshTokenRepository) Create(ctx context.Context, userID string, token string, expiresAt time.Time) (*RefreshToken, error) {
	now := time.Now()
	entity := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     hashToken(token),
		ExpiresAt: expiresAt,
		Revoked:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := database.GetDB(ctx, r.db).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*RefreshToken, error) {
	var entity RefreshToken
	if err := database.GetDB(ctx, r.db).First(&entity, "token = ?", hashToken(token)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

func (r *GormRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	res := database.GetDB(ctx, r.db).Model(&RefreshToken{}).Where("token = ? AND revoked = ?", hashToken(token), false).Updates(map[string]interface{}{
		"revoked":    true,
		"updated_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrUnauthorized
	}
	return nil
}

func (r *GormRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	return database.GetDB(ctx, r.db).Model(&RefreshToken{}).Where("user_id = ? AND revoked = ?", userID, false).Updates(map[string]interface{}{
		"revoked":    true,
		"updated_at": time.Now(),
	}).Error
}

func (r *GormRefreshTokenRepository) RotateToken(ctx context.Context, oldToken string, userID string, newToken string, expiresAt time.Time) (*RefreshToken, error) {
	var newEntity *RefreshToken
	now := time.Now()
	err := database.WithTransaction(ctx, r.db, func(txCtx context.Context) error {
		tx := database.GetDB(txCtx, r.db)
		res := tx.Model(&RefreshToken{}).Where("token = ? AND revoked = ?", hashToken(oldToken), false).Updates(map[string]interface{}{
			"revoked":    true,
			"updated_at": now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return common.ErrUnauthorized
		}

		newEntity = &RefreshToken{
			ID:        uuid.New().String(),
			UserID:    userID,
			Token:     hashToken(newToken),
			ExpiresAt: expiresAt,
			Revoked:   false,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return tx.Create(newEntity).Error
	})
	if err != nil {
		return nil, err
	}
	return newEntity, nil
}

func (r *GormRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res := database.GetDB(ctx, r.db).Where("expires_at < ?", time.Now()).Delete(&RefreshToken{})
	return res.RowsAffected, res.Error
}
