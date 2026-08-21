package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/danarrigo/scaean-gate/applications/app-a/config"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/dto"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/models"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/pkg/crypto"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/repository"
	"github.com/google/uuid"
)

type AuthService struct {
	Repo       *repository.SessionRepository
	Cfg        *config.Config
	HTTPClient *http.Client
}

func NewAuthService(repo *repository.SessionRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		Repo: repo,
		Cfg:  cfg,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *AuthService) GenerateLoginURL() (string, string, string, error) {
	verifier, challenge, state, err := crypto.GeneratePKCE()
	if err != nil {
		return "", "", "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate PKCE challenge")
	}

	authURL := fmt.Sprintf(
		"%s/authorize?client_id=%s&redirect_uri=%s&response_type=code&code_challenge=%s&code_challenge_method=S256&state=%s",
		s.Cfg.AuthProviderURL,
		s.Cfg.ClientID,
		url.QueryEscape(s.Cfg.RedirectURI),
		challenge,
		state,
	)

	return authURL, verifier, state, nil
}

func (s *AuthService) HandleCallback(ctx context.Context, code, verifier string) (string, error) {
	tokenReqBody := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  s.Cfg.RedirectURI,
		"client_id":     s.Cfg.ClientID,
		"client_secret": s.Cfg.ClientSecret,
		"code_verifier": verifier,
	}

	jsonBody, err := json.Marshal(tokenReqBody)
	if err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to prepare token request")
	}

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/token", s.Cfg.AuthProviderInternalURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to build token request")
	}
	tokenReq.Header.Set("Content-Type", "application/json")

	tokenResp, err := s.HTTPClient.Do(tokenReq)
	if err != nil {
		return "", response.NewAppError(http.StatusBadGateway, "SSO_UNAVAILABLE", "Failed to connect to Auth Provider")
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return "", response.NewAppError(http.StatusUnauthorized, "TOKEN_EXCHANGE_FAILED", "Failed to exchange authorization code for token")
	}

	var tokenData dto.TokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to parse token response")
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/userinfo", s.Cfg.AuthProviderInternalURL), nil)
	if err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to build userinfo request")
	}
	userReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.AccessToken))

	userResp, err := s.HTTPClient.Do(userReq)
	if err != nil {
		return "", response.NewAppError(http.StatusBadGateway, "SSO_UNAVAILABLE", "Failed to fetch user claims")
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		return "", response.NewAppError(http.StatusUnauthorized, "USERINFO_FAILED", "Failed to fetch user profile claims")
	}

	var userInfo dto.UserInfoResponse
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to parse userinfo response")
	}

	userUUID, err := uuid.Parse(userInfo.Sub)
	if err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid user ID format from SSO")
	}

	now := time.Now()
	profile := models.ProfileCache{
		ExternalUserID: userUUID,
		Name:           userInfo.Name,
		Email:          userInfo.Email,
		Groups:         models.StringArray(userInfo.Groups),
		SyncedAt:       now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.Repo.UpsertProfileCache(&profile); err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cache user profile")
	}

	rawLocalToken, err := crypto.GenerateRandomString(32)
	if err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate local session token")
	}

	var centralSessionID uuid.UUID
	if tokenData.CentralSessionID != nil {
		centralSessionID = *tokenData.CentralSessionID
	}

	localSession := models.LocalSession{
		SessionTokenHash: crypto.HashSHA256(rawLocalToken),
		ExternalUserID:   userUUID,
		CentralSessionID: centralSessionID,
		Status:           "active",
		CreatedAt:        now,
		ExpiresAt:        now.Add(24 * time.Hour),
	}
	if err := s.Repo.CreateLocalSession(&localSession); err != nil {
		return "", response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create local session")
	}

	return rawLocalToken, nil
}

func (s *AuthService) GetMe(session *models.LocalSession) (*dto.MeResponse, error) {
	profile, err := s.Repo.GetProfileCache(session.ExternalUserID)
	if err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "User profile not found in cache")
	}

	return &dto.MeResponse{
		AppName: s.Cfg.AppName,
		User: dto.UserProfileDTO{
			ID:     profile.ExternalUserID,
			Name:   profile.Name,
			Email:  profile.Email,
			Groups: []string(profile.Groups),
		},
		Session: dto.SessionDTO{
			ID:        session.ID,
			Status:    session.Status,
			CreatedAt: session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
		},
	}, nil
}

func (s *AuthService) Logout(sessionID uuid.UUID) error {
	return s.Repo.RevokeLocalSession(sessionID, "Local Logout")
}

func (s *AuthService) HandleBackChannelLogout(payload dto.BackChannelLogoutPayload) error {
	isProcessed, err := s.Repo.IsEventProcessed(payload.EventID)
	if err != nil {
		return response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check idempotency")
	}
	if isProcessed {
		return nil
	}

	if payload.CentralSessionID != nil && *payload.CentralSessionID != uuid.Nil {
		if err := s.Repo.RevokeLocalSessionsByCentralSessionID(*payload.CentralSessionID, payload.Reason); err != nil {
			return response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke local sessions")
		}
	} else if payload.UserID != uuid.Nil {
		if err := s.Repo.RevokeLocalSessionsByUserID(payload.UserID, payload.Reason); err != nil {
			return response.NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke local sessions")
		}
	} else {
		return response.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "Logout event has no session or user target")
	}

	processedEvent := models.ProcessedEvent{
		EventID:     payload.EventID,
		EventType:   payload.EventType,
		ProcessedAt: time.Now(),
		Result:      "success",
	}
	return s.Repo.RecordProcessedEvent(&processedEvent)
}

func (s *AuthService) GetProcessedEvents() ([]models.ProcessedEvent, error) {
	return s.Repo.GetProcessedEvents(20)
}
