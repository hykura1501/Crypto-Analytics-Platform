package repository

import (
	"time"

	"github.com/crypto-platform/auth-service/internal/model"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *model.RefreshToken) error
	FindByToken(token string) (*model.RefreshToken, error)
	DeleteByToken(token string) error
	DeleteByUserID(userID uint) error
	DeleteExpired() error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByToken(token string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).
		Preload("User").
		First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&model.RefreshToken{}).Error
}
