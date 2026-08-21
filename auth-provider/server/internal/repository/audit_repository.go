package repository

import (
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository struct {
	DB *gorm.DB
}

func (r *AuditRepository) CreateAuditLog(auditLog *models.AuditLog) error {
	return r.DB.Create(auditLog).Error
}

func (r *AuditRepository) ListAuditLogs() ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.DB.Order("created_at desc").Limit(100).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AuditRepository) ListEvents() ([]models.Event, error) {
	var events []models.Event
	if err := r.DB.Preload("Deliveries.Application").Order("created_at desc").Limit(100).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *AuditRepository) MarkEventAsFinished(eventID uuid.UUID) error {
	now := time.Now()
	return r.DB.Model(&models.Event{}).Where("id = ?", eventID).Updates(map[string]any{
		"status":       "published",
		"published_at": &now,
	}).Error
}
