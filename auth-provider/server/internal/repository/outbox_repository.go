// Package repository contains CRUD ops
package repository

import (
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"gorm.io/gorm"
)

type OutboxRepository struct {
	DB *gorm.DB
}

func (r *OutboxRepository) CreateEvent(event *models.Event) error {
	return r.DB.Create(event).Error
}
