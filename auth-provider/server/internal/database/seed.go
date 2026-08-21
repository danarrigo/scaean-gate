package database

import (
	"fmt"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	appcrypto "github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/crypto"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB, seedPassword, appAClientSecret, appBClientSecret, appALogoutURL, appBLogoutURL string) error {
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
		Name:                  "Apex",
		ClientID:              "app-a-client-id",
		ClientSecretHash:      appcrypto.HashSHA256(appAClientSecret),
		ClientSecretPrefix:    secretPrefix(appAClientSecret),
		Status:                "active",
		LaunchURL:             "http://localhost:4201",
		LogoutNotificationURL: appALogoutURL,
	}

	appB := models.Application{
		Name:                  "Bolt",
		ClientID:              "app-b-client-id",
		ClientSecretHash:      appcrypto.HashSHA256(appBClientSecret),
		ClientSecretPrefix:    secretPrefix(appBClientSecret),
		Status:                "active",
		LaunchURL:             "http://localhost:4202",
		LogoutNotificationURL: appBLogoutURL,
	}

	err = db.Where(models.Application{ClientID: appA.ClientID}).
		Assign(models.Application{ClientSecretHash: appA.ClientSecretHash, ClientSecretPrefix: appA.ClientSecretPrefix, LogoutNotificationURL: appA.LogoutNotificationURL}).
		FirstOrCreate(&appA).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.Where(models.Application{ClientID: appB.ClientID}).
		Assign(models.Application{ClientSecretHash: appB.ClientSecretHash, ClientSecretPrefix: appB.ClientSecretPrefix, LogoutNotificationURL: appB.LogoutNotificationURL}).
		FirstOrCreate(&appB).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	uris := []models.ApplicationRedirectURI{
		{ApplicationID: appA.ID, RedirectURI: "http://localhost:8081/auth/callback"},
		{ApplicationID: appA.ID, RedirectURI: "http://localhost:4201/callback"},
		{ApplicationID: appB.ID, RedirectURI: "http://localhost:8082/auth/callback"},
		{ApplicationID: appB.ID, RedirectURI: "http://localhost:4202/callback"},
	}

	for _, u := range uris {
		err = db.FirstOrCreate(&u, models.ApplicationRedirectURI{
			ApplicationID: u.ApplicationID,
			RedirectURI:   u.RedirectURI,
		}).Error
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}
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

func secretPrefix(secret string) string {
	if len(secret) <= 11 {
		return secret
	}
	return secret[:11]
}
