// Package database Handles Connection and Migration
package database

import (
	"fmt"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if err := db.Exec(`DROP INDEX IF EXISTS "idx_user_email_deleted"`).Error; err != nil {
		return fmt.Errorf("drop obsolete user email index: %w", err)
	}
	if err := db.SetupJoinTable(&models.User{}, "Groups", &models.UserGroup{}); err != nil {
		return fmt.Errorf("configure user-group join table: %w", err)
	}
	if err := db.SetupJoinTable(&models.Group{}, "Users", &models.UserGroup{}); err != nil {
		return fmt.Errorf("configure group-user join table: %w", err)
	}

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
