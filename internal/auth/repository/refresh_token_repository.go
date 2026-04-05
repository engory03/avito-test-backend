package repository

import (
	"context"

	"github.com/internships-backend/test-backend-engory03/internal/auth/model"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, token *model.RefreshToken) error
	Delete(ctx context.Context, token string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Save(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}
