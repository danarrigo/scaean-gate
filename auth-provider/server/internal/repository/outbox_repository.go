// Package repository contains CRUD ops
package repository

import (
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxRepository struct {
	DB *gorm.DB
}

func (r *OutboxRepository) CreateEvent(event *models.Event) error {
	return r.DB.Create(event).Error
}

func (r *OutboxRepository) ListPendingEvents(limit int) ([]models.Event, error) {
	var events []models.Event
	err := r.DB.Where("status = ?", "pending").Order("created_at ASC").Limit(limit).Find(&events).Error
	return events, err
}

func (r *OutboxRepository) MarkPublished(eventID uuid.UUID) error {
	now := time.Now()
	return r.DB.Model(&models.Event{}).
		Where("id = ? AND status = ?", eventID, "pending").
		Updates(map[string]any{"status": "published", "published_at": &now}).Error
}
