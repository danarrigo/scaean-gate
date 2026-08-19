package repository

import (
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
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
	if err := r.DB.Order("created_at desc").Limit(100).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
