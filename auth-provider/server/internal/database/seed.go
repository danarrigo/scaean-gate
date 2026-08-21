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

	err := db.Unscoped().Where("name = ?", adminGroup.Name).FirstOrCreate(&adminGroup).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	err = db.Unscoped().Where("name = ?", userGroup.Name).FirstOrCreate(&userGroup).Error
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

	err = db.Unscoped().Where("email = ?", adminUser.Email).FirstOrCreate(&adminUser).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.Unscoped().Where("email = ?", testUser.Email).FirstOrCreate(&testUser).Error
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

	err = db.Unscoped().Where("user_id = ? AND group_id = ?", adminUG.UserID, adminUG.GroupID).FirstOrCreate(&adminUG).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.Unscoped().Where("user_id = ? AND group_id = ?", testUG.UserID, testUG.GroupID).FirstOrCreate(&testUG).Error
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

	err = db.Unscoped().Where("client_id = ?", appA.ClientID).
		Assign(map[string]any{"client_secret_hash": appA.ClientSecretHash, "client_secret_prefix": appA.ClientSecretPrefix, "logout_notification_url": appA.LogoutNotificationURL}).
		FirstOrCreate(&appA).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.Unscoped().Where("client_id = ?", appB.ClientID).
		Assign(map[string]any{"client_secret_hash": appB.ClientSecretHash, "client_secret_prefix": appB.ClientSecretPrefix, "logout_notification_url": appB.LogoutNotificationURL}).
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
		err = db.Unscoped().Where("application_id = ? AND redirect_uri = ?", u.ApplicationID, u.RedirectURI).FirstOrCreate(&u).Error
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

	err = db.Unscoped().Where("application_id = ? AND group_id = ? AND effect = 'allow'", appAPolicy.ApplicationID, appAPolicy.GroupID).FirstOrCreate(&appAPolicy).Error
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	err = db.Unscoped().Where("application_id = ? AND group_id = ? AND effect = 'allow'", appBPolicy.ApplicationID, appBPolicy.GroupID).FirstOrCreate(&appBPolicy).Error
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
