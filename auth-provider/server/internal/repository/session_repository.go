package repository

import (
	"errors"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	
	"github.com/google/uuid"
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

func (r *SessionRepository) GetSSOSessionByHash (hash string)(models.SSOSession,error){
	ssoSession := make ([]models.SSOSession,0)
	if err:= r.DB.Where("session_token_hash = ?",hash).Find(&ssoSession).Error;err!=nil{
		return models.SSOSession{},err
	}
	if len (ssoSession) < 1{
		return models.SSOSession{},errors.New("no session found")
	}
	if len (ssoSession) > 1 {
		return models.SSOSession{},errors.New("multiple sessions found")
	}
	return ssoSession[0],nil
}

func (r *SessionRepository) RevokeSSOSession(id uuid.UUID) error {
	if err := r.DB.Model(&models.SSOSession{}).Where("id = ?", id).Updates(map[string]any{
		"status":        "revoked",
		"revoked_at":    time.Now(),
		"revoke_reason": "USER_LOGOUT",
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *SessionRepository) RevokeSSOSessionWithEvent(id uuid.UUID, reason string, event *models.Event) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.SSOSession{}).Where("id = ?", id).Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    time.Now(),
			"revoke_reason": reason,
		}).Error; err != nil {
			return err
		}

		if event != nil {
			if err := tx.Create(event).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
