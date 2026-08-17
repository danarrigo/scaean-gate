// Package services implements application logic
package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type LoginResult struct {
	User     models.User
	RawToken string
}

type AuthService struct {
	UserRepo    repository.UserRepository
	SessionRepo repository.SessionRepository
}

func (s *AuthService) Login(email, password, ipAddress, userAgent string) (*LoginResult, error) {
	user, err := s.UserRepo.GetUserByEmail(email)
	if err != nil {
		return nil, response.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, response.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	}

	if user.Status != "active" {
		return nil, response.NewAppError(http.StatusForbidden, "USER_INACTIVE", "User account is inactive")
	}

	rawToken, err := s.GenerateRandomString()
	if err != nil {
		return nil, err
	}

	sessionTokenHash := s.GenerateSessionToken(rawToken)

	session := models.SSOSession{
		UserID:           user.ID,
		SessionTokenHash: sessionTokenHash,
		Status:           "active",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
	}

	if err := s.SessionRepo.CreateSSOSession(&session); err != nil {
		return nil, err
	}

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

func (s *AuthService) GenerateRandomString() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	randomString := hex.EncodeToString(bytes)
	return randomString, nil
}

func (s *AuthService) GenerateSessionToken(rawToken string) string {
	hashedBytes := sha256.Sum256([]byte(rawToken))
	hashedString := hex.EncodeToString(hashedBytes[:])
	return hashedString
}

func (s *AuthService) CreateSession(session *models.SSOSession) error {
	return s.SessionRepo.CreateSSOSession(session)
}
