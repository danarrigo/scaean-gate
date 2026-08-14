package database

import (
	"fmt"
	"os"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	adminGroup := models.Group{
		Name:        "Admin",
		Description: "Group for Administrators",
	}
	userGroup := models.Group{
		Name:        "User",
		Description: "Group for Users",
	}

	err := db.FirstOrCreate(&adminGroup, models.Group{Name: adminGroup.Name}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	err = db.FirstOrCreate(&userGroup, models.Group{Name: userGroup.Name}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	seedPassword := os.Getenv("SEED_USER_PASSWORD")
	if seedPassword == "" {
		return fmt.Errorf("missing required environment variable: SEED_USER_PASSWORD")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	adminUser := models.User{
		Name:         "Admin User",
		Email:        "admin@scaean-gate.com",
		PasswordHash: string(hashedPassword),
		Status:       "active",
	}

	testUser := models.User{
		Name:         "Test User",
		Email:        "testuser@scaean-gate.com",
		PasswordHash: string(hashedPassword),
		Status:       "active",
	}

	err = db.FirstOrCreate(&adminUser, models.User{Email: adminUser.Email}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.FirstOrCreate(&testUser, models.User{Email: testUser.Email}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	adminUG := models.UserGroup{
		GroupID: adminGroup.ID,
		UserID:  adminUser.ID,
	}

	testUG := models.UserGroup{
		UserID:  testUser.ID,
		GroupID: userGroup.ID,
	}

	err = db.FirstOrCreate(&adminUG, models.UserGroup{
		UserID:  adminUG.UserID,
		GroupID: adminUG.GroupID,
	}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.FirstOrCreate(&testUG, models.UserGroup{
		UserID:  testUG.UserID,
		GroupID: testUG.GroupID,
	}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	appA := models.Application{
		Name:                  "App A",
		ClientID:              "app-a-client-id",
		ClientSecretHash:      string(hashedPassword),
		Status:                "active",
		LaunchURL:             "http://localhost:4201",
		LogoutNotificationURL: "http://localhost:8081/internal/logout",
	}

	appB := models.Application{
		Name:                  "App B",
		ClientID:              "app-b-client-id",
		ClientSecretHash:      string(hashedPassword),
		Status:                "active",
		LaunchURL:             "http://localhost:4202",
		LogoutNotificationURL: "http://localhost:8082/internal/logout",
	}

	err = db.FirstOrCreate(&appA, models.Application{ClientID: appA.ClientID}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.FirstOrCreate(&appB, models.Application{ClientID: appB.ClientID}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	appARedirect := models.ApplicationRedirectURI{
		ApplicationID: appA.ID,
		RedirectURI:   "http://localhost:4201/callback",
	}

	appBRedirect := models.ApplicationRedirectURI{
		ApplicationID: appB.ID,
		RedirectURI:   "http://localhost:4202/callback",
	}

	err = db.FirstOrCreate(&appARedirect, models.ApplicationRedirectURI{
		ApplicationID: appARedirect.ApplicationID,
		RedirectURI:   appARedirect.RedirectURI,
	}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.FirstOrCreate(&appBRedirect, models.ApplicationRedirectURI{
		ApplicationID: appBRedirect.ApplicationID,
		RedirectURI:   appBRedirect.RedirectURI,
	}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	appAPolicy := models.ApplicationGroupPolicy{
		ApplicationID: appA.ID,
		GroupID:       userGroup.ID,
		Effect:        "allow",
	}

	appBPolicy := models.ApplicationGroupPolicy{
		ApplicationID: appB.ID,
		GroupID:       userGroup.ID,
		Effect:        "allow",
	}

	err = db.FirstOrCreate(&appAPolicy, models.ApplicationGroupPolicy{
		ApplicationID: appAPolicy.ApplicationID,
		GroupID:       appAPolicy.GroupID,
		Effect:        "allow",
	}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.FirstOrCreate(&appBPolicy, models.ApplicationGroupPolicy{
		ApplicationID: appBPolicy.ApplicationID,
		GroupID:       appBPolicy.GroupID,
		Effect:        "allow",
	}).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}
