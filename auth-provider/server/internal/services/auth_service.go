// Package services implements application logic
package services

import (
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/crypto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginResult struct {
	User     models.User
	RawToken string
}

type AuthService struct {
	UserRepo    repository.UserRepository
	SessionRepo repository.SessionRepository
	AuditSvc    AuditService
}

func (s *AuthService) Login(email, password, ipAddress, userAgent string) (*LoginResult, error) {
	var (
		userID    *uuid.UUID
		sessionID *uuid.UUID
		result    = "failed"
		reason    = ""
	)

	defer func() {
		s.AuditSvc.Log("LOGIN", userID, userID, nil, sessionID, result, ipAddress, reason)
	}()

	user, err := s.UserRepo.GetUserByEmail(email)
	if err != nil {
		reason = "INVALID_CREDENTIALS"
		return nil, response.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	}

	userID = &user.ID

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		reason = "INVALID_CREDENTIALS"
		return nil, response.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	}

	if user.Status != "active" {
		reason = "USER_INACTIVE"
		return nil, response.NewAppError(http.StatusForbidden, "USER_INACTIVE", "User account is inactive")
	}

	rawToken, err := crypto.GenerateRandomString()
	if err != nil {
		reason = "INTERNAL_ERROR"
		return nil, err
	}

	sessionTokenHash := crypto.HashSHA256(rawToken)

	session := models.SSOSession{
		UserID:           user.ID,
		SessionTokenHash: sessionTokenHash,
		Status:           "active",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
	}

	if err := s.SessionRepo.CreateSSOSession(&session); err != nil {
		reason = "INTERNAL_ERROR"
		return nil, err
	}

	sessionID = &session.ID
	result = "success"

	return &LoginResult{
		User:     user,
		RawToken: rawToken,
	}, nil
}

func (s *AuthService) GetUserByEmail(email string) (models.User, error) {
	return s.UserRepo.GetUserByEmail(email)
}

func (s *AuthService) ValidateUserPassword(hashedPassword string, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) CreateSession(session *models.SSOSession) error {
	return s.SessionRepo.CreateSSOSession(session)
}
