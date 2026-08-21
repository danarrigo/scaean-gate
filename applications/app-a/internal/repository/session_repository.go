package repository

import (
	"time"

	"github.com/danarrigo/scaean-gate/applications/app-a/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SessionRepository struct {
	DB *gorm.DB
}

func (r *SessionRepository) CreateLocalSession(session *models.LocalSession) error {
	return r.DB.Create(session).Error
}

func (r *SessionRepository) GetLocalSessionByTokenHash(tokenHash string) (*models.LocalSession, error) {
	var session models.LocalSession
	err := r.DB.Where("session_token_hash = ?", tokenHash).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) RevokeLocalSession(sessionID uuid.UUID, reason string) error {
	now := time.Now()
	return r.DB.Model(&models.LocalSession{}).
		Where("id = ? AND status = 'active'", sessionID).
		Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    &now,
			"revoke_reason": reason,
		}).Error
}

func (r *SessionRepository) RevokeLocalSessionsByCentralSessionID(centralSessionID uuid.UUID, reason string) error {
	now := time.Now()
	return r.DB.Model(&models.LocalSession{}).
		Where("central_session_id = ? AND status = 'active'", centralSessionID).
		Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    &now,
			"revoke_reason": reason,
		}).Error
}

func (r *SessionRepository) RevokeLocalSessionsByUserID(userID uuid.UUID, reason string) error {
	now := time.Now()
	return r.DB.Model(&models.LocalSession{}).
		Where("external_user_id = ? AND status = 'active'", userID).
		Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    &now,
			"revoke_reason": reason,
		}).Error
}

func (r *SessionRepository) UpsertProfileCache(profile *models.ProfileCache) error {
	return r.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "email", "groups", "synced_at", "updated_at"}),
	}).Create(profile).Error
}

func (r *SessionRepository) GetProfileCache(userID uuid.UUID) (*models.ProfileCache, error) {
	var profile models.ProfileCache
	err := r.DB.Where("external_user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *SessionRepository) IsEventProcessed(eventID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.Model(&models.ProcessedEvent{}).Where("event_id = ?", eventID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SessionRepository) RecordProcessedEvent(event *models.ProcessedEvent) error {
	return r.DB.Create(event).Error
}

func (r *SessionRepository) GetProcessedEvents(limit int) ([]models.ProcessedEvent, error) {
	var events []models.ProcessedEvent
	err := r.DB.Order("processed_at desc").Limit(limit).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *SessionRepository) RecordAuthActivity(activity *models.AuthActivity) error {
	return r.DB.Create(activity).Error
}

func (r *SessionRepository) GetAuthActivities(flowID string, sessionID uuid.UUID) ([]models.AuthActivity, error) {
	var activities []models.AuthActivity
	if err := r.DB.Select("id, oauth_flow_id, event_type, result, created_at").
		Where("oauth_flow_id = ? OR local_session_id = ?", flowID, sessionID).
		Order("created_at asc").Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}
