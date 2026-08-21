package repository

import (
	"errors"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrAuthorizationCodeConsumed = errors.New("authorization code is expired or already used")

type OAuthRepository struct {
	DB *gorm.DB
}

func (r *OAuthRepository) CreateAuthorizationCode(code *models.AuthorizationCode) error {
	return r.DB.Create(code).Error
}

func (r *OAuthRepository) GetAuthCodeByHash(codeHash string) (models.AuthorizationCode, error) {
	var code models.AuthorizationCode
	if err := r.DB.Where("code_hash = ?", codeHash).First(&code).Error; err != nil {
		return models.AuthorizationCode{}, err
	}
	return code, nil
}

func (r *OAuthRepository) ConsumeAuthorizationCode(id uuid.UUID, token *models.AccessToken) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&models.AuthorizationCode{}).
			Where("id = ? AND used_at IS NULL AND expires_at > ?", id, now).
			Update("used_at", &now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrAuthorizationCodeConsumed
		}
		return tx.Create(token).Error
	})
}

func (r *OAuthRepository) GetAccessTokenByHash(tokenHash string) (models.AccessToken, error) {
	var token models.AccessToken
	if err := r.DB.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		return models.AccessToken{}, err
	}
	return token, nil
}
