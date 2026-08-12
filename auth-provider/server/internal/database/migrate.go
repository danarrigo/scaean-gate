// Package database Handles Connection and Migration
package database

import (
	"fmt"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.UserGroup{},
		&models.Application{},
		&models.ApplicationRedirectURI{},
		&models.ApplicationGroupPolicy{},
		&models.SSOSession{},
		&models.AuthorizationCode{},
		&models.AccessToken{},
		&models.AuditLog{},
		&models.Event{},
		&models.EventDelivery{},
	)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	fmt.Printf("Database Migration Succeeded")
	return nil
}
