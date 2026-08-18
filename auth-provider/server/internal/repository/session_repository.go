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
	return r.DB.Create(session).Error
}

func (r *SessionRepository) GetSSOSessionByID(id uuid.UUID) (models.SSOSession, error) {
	var session models.SSOSession
	if err := r.DB.Where("id = ?", id).First(&session).Error; err != nil {
		return models.SSOSession{}, err
	}
	return session, nil
}

func (r *SessionRepository) GetSSOSessionByHash(hash string) (models.SSOSession, error) {
	var sessions []models.SSOSession
	if err := r.DB.Where("session_token_hash = ?", hash).Find(&sessions).Error; err != nil {
		return models.SSOSession{}, err
	}
	if len(sessions) < 1 {
		return models.SSOSession{}, errors.New("no session found")
	}
	if len(sessions) > 1 {
		return models.SSOSession{}, errors.New("multiple sessions found")
	}
	return sessions[0], nil
}

func (r *SessionRepository) RevokeSSOSession(id uuid.UUID) error {
	return r.DB.Model(&models.SSOSession{}).Where("id = ?", id).Updates(map[string]any{
		"status":        "revoked",
		"revoked_at":    time.Now(),
		"revoke_reason": "USER_LOGOUT",
	}).Error
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
