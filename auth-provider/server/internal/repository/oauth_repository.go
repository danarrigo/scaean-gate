package repository

import (
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func (r *OAuthRepository) MarkAuthCodeUsed(id uuid.UUID) error {
	now := time.Now()
	return r.DB.Model(&models.AuthorizationCode{}).Where("id = ?", id).Update("used_at", &now).Error
}

func (r *OAuthRepository) CreateAccessToken(token *models.AccessToken) error {
	return r.DB.Create(token).Error
}

func (r *OAuthRepository) GetAccessTokenByHash(tokenHash string) (models.AccessToken, error) {
	var token models.AccessToken
	if err := r.DB.Where("token_hash = ?", tokenHash).First(&token).Error; err != nil {
		return models.AccessToken{}, err
	}
	return token, nil
}
