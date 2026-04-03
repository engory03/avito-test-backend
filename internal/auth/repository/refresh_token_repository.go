package repository

import (
	"github.com/internships-backend/test-backend-engory03/internal/auth/model"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Save(token *model.RefreshToken) error
	Delete(token string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Save(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}
