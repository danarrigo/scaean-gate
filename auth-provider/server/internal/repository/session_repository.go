package repository

import (
	
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"

	"gorm.io/gorm"
)

type SessionRepository struct {
	DB *gorm.DB
}

func (r *SessionRepository) CreateSSOSession(session *models.SSOSession) error {
	if err := r.DB.Create(session).Error; err != nil {
		return err
	}
	return nil
}
