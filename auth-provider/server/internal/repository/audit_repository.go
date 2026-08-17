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
