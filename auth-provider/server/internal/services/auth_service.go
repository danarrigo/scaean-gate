// Package services implements application logic
package services

import (
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
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

func (s *AuthService) Logout(cookie, ipAddress string) error {
	hashedToken := crypto.HashSHA256(cookie)
	ssoSession, err := s.SessionRepo.GetSSOSessionByHash(hashedToken)
	if err != nil {
		return response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "Invalid Session")
	}

	event := models.Event{
		EventType:        "SessionRevoked",
		UserID:           ssoSession.UserID,
		CentralSessionID: &ssoSession.ID,
		Payload:          `{"reason":"USER_LOGOUT"}`,
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	if err := s.SessionRepo.RevokeSSOSessionWithEvent(ssoSession.ID, "USER_LOGOUT", &event); err != nil {
		return err
	}

	s.AuditSvc.Log("LOGOUT", &ssoSession.UserID, &ssoSession.UserID, nil, &ssoSession.ID, "success", ipAddress, "")
	return nil
}

func (s *AuthService) ChangePassword(cookie string, req dto.ChangePasswordRequest, ipAddress string) error {
	var (
		userID *uuid.UUID
		result = "failed"
		reason = ""
	)

	defer func() {
		s.AuditSvc.Log("PASSWORD_CHANGED", userID, userID, nil, nil, result, ipAddress, reason)
	}()

	hashedToken := crypto.HashSHA256(cookie)
	session, err := s.SessionRepo.GetSSOSessionByHash(hashedToken)
	if err != nil {
		reason = "UNAUTHORIZED"
		return response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired session")
	}

	userID = &session.UserID

	user, err := s.UserRepo.GetUserByID(session.UserID)
	if err != nil {
		reason = "USER_NOT_FOUND"
		return response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "User not found")
	}

	if err := crypto.ValidateUserPassword(req.OldPassword, user.PasswordHash); err != nil {
		reason = "INVALID_CREDENTIALS"
		return response.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect old password")
	}

	newPasswordHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		reason = "INTERNAL_ERROR"
		return err
	}

	event := models.Event{
		EventType: "PasswordChanged",
		UserID:    user.ID,
		Payload:   `{"reason":"PASSWORD_CHANGED"}`,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := s.UserRepo.ChangePasswordTx(user.ID, newPasswordHash, &event); err != nil {
		reason = "INTERNAL_ERROR"
		return err
	}

	result = "success"
	return nil
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

