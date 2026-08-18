package services

import (
	"net/http"
	"net/url"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/crypto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
)

type OauthService struct {
	SessionRepo repository.SessionRepository
	UserRepo    repository.UserRepository
	AppRepo     repository.AppRepository
	PolicyRepo  repository.PolicyRepository
	OAuthRepo   repository.OAuthRepository
}

func (s *OauthService) AuthorizeService(rawToken string, req dto.AuthorizeRequest) (string, error) {
	session, err := s.SessionRepo.GetSSOSessionByHash(crypto.HashSHA256(rawToken))
	if err != nil || session.Status != "active" || session.ExpiresAt.Before(time.Now()) {
		return "", response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "Invalid Session")
	}

	user, err := s.UserRepo.GetUserByID(session.UserID)
	if err != nil || user.Status != "active" {
		return "", response.NewAppError(http.StatusForbidden, "FORBIDDEN", "User Inactive")
	}

	app, err := s.AppRepo.GetAppByID(req.ClientID)
	if err != nil || app.Status != "active" {
		return "", response.NewAppError(http.StatusBadRequest, "INVALID_CLIENT", "Invalid Client")
	}

	uriMatch := false
	for _, uri := range app.RedirectURIs {
		if uri.RedirectURI == req.RedirectURI {
			uriMatch = true
			break
		}
	}

	if !uriMatch {
		return "", response.NewAppError(http.StatusBadRequest, "INVALID_REDIRECT_URL", "Invalid Redirect URI")
	}

	hasAccess, err := s.PolicyRepo.HasUserAccess(user.ID, app.ID)
	if err != nil || !hasAccess {
		return "", response.NewAppError(http.StatusForbidden, "ACCESS_DENIED", "Access Denied by Policy")
	}

	rawCode, err := crypto.GenerateRandomString()
	if err != nil {
		return "", err
	}

	authCode := models.AuthorizationCode{
		CodeHash:            crypto.HashSHA256(rawCode),
		UserID:              user.ID,
		ApplicationID:       app.ID,
		SSOSessionID:        session.ID,
		RedirectURI:         req.RedirectURI,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(3 * time.Minute),
	}

	if err := s.OAuthRepo.CreateAuthorizationCode(&authCode); err != nil {
		return "", err
	}

	targetURL, err := url.Parse(req.RedirectURI)
	if err != nil {
		return "", response.NewAppError(http.StatusBadRequest, "INVALID_REDIRECT_URL", "Invalid Redirect URI")
	}

	q := targetURL.Query()
	q.Set("code", rawCode)
	if req.State != "" {
		q.Set("state", req.State)
	}
	targetURL.RawQuery = q.Encode()

	return targetURL.String(), nil
}

func (s *OauthService) ExchangeToken(req dto.TokenRequest) (*dto.TokenResponse, error) {
	codeHash := crypto.HashSHA256(req.Code)
	authCode, err := s.OAuthRepo.GetAuthCodeByHash(codeHash)
	if err != nil || authCode.UsedAt != nil || authCode.ExpiresAt.Before(time.Now()) {
		return nil, response.NewAppError(http.StatusBadRequest, "INVALID_GRANT", "Invalid, expired, or already used authorization code")
	}

	app, err := s.AppRepo.GetAppByID(req.ClientID)
	if err != nil || app.ID != authCode.ApplicationID || app.Status != "active" {
		return nil, response.NewAppError(http.StatusBadRequest, "INVALID_GRANT", "Invalid client")
	}

	if authCode.RedirectURI != req.RedirectURI {
		return nil, response.NewAppError(http.StatusBadRequest, "INVALID_GRANT", "Redirect URI mismatch")
	}

	if !crypto.VerifyPKCE(req.CodeVerifier, authCode.CodeChallenge) {
		return nil, response.NewAppError(http.StatusBadRequest, "INVALID_GRANT", "Invalid PKCE code verifier")
	}

	session, err := s.SessionRepo.GetSSOSessionByID(authCode.SSOSessionID)
	if err != nil || session.Status != "active" || session.ExpiresAt.Before(time.Now()) {
		return nil, response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "SSO session is no longer active")
	}

	if err := s.OAuthRepo.MarkAuthCodeUsed(authCode.ID); err != nil {
		return nil, err
	}

	rawToken, err := crypto.GenerateRandomString()
	if err != nil {
		return nil, err
	}

	accessToken := models.AccessToken{
		TokenHash:     crypto.HashSHA256(rawToken),
		UserID:        authCode.UserID,
		ApplicationID: authCode.ApplicationID,
		SSOSessionID:  authCode.SSOSessionID,
		Status:        "active",
		IssuedAt:      time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	if err := s.OAuthRepo.CreateAccessToken(&accessToken); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken: rawToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}, nil
}

func (s *OauthService) GetUserInfo(rawAccessToken string) (*dto.UserInfoResponse, error) {
	tokenHash := crypto.HashSHA256(rawAccessToken)
	accessToken, err := s.OAuthRepo.GetAccessTokenByHash(tokenHash)
	if err != nil || accessToken.Status != "active" || accessToken.ExpiresAt.Before(time.Now()) {
		return nil, response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token")
	}

	session, err := s.SessionRepo.GetSSOSessionByID(accessToken.SSOSessionID)
	if err != nil || session.Status != "active" || session.ExpiresAt.Before(time.Now()) {
		return nil, response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "Associated SSO session is inactive")
	}

	user, err := s.UserRepo.GetUserWithGroupsByID(accessToken.UserID)
	if err != nil || user.Status != "active" {
		return nil, response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "User is inactive or not found")
	}

	groups := make([]string, 0, len(user.Groups))
	for _, g := range user.Groups {
		groups = append(groups, g.Name)
	}

	return &dto.UserInfoResponse{
		Sub:    user.ID.String(),
		Name:   user.Name,
		Email:  user.Email,
		Groups: groups,
	}, nil
}
