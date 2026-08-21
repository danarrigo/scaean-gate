package dto

import "github.com/google/uuid"

type AuthorizeRequest struct {
	ClientID            string `form:"client_id" binding:"required"`
	RedirectURI         string `form:"redirect_uri" binding:"required,url"`
	ResponseType        string `form:"response_type" binding:"required,eq=code"`
	State               string `form:"state"`
	CodeChallenge       string `form:"code_challenge" binding:"required"`
	CodeChallengeMethod string `form:"code_challenge_method" binding:"required,eq=S256"`
}

type TokenRequest struct {
	GrantType    string `form:"grant_type" json:"grant_type" binding:"required,eq=authorization_code"`
	Code         string `form:"code" json:"code" binding:"required"`
	RedirectURI  string `form:"redirect_uri" json:"redirect_uri" binding:"required,url"`
	ClientID     string `form:"client_id" json:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" json:"client_secret"`
	CodeVerifier string `form:"code_verifier" json:"code_verifier" binding:"required"`
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	TokenType        string    `json:"token_type"`
	ExpiresIn        int       `json:"expires_in"`
	CentralSessionID uuid.UUID `json:"central_session_id"`
}

type UserInfoResponse struct {
	Sub    string   `json:"sub"`
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}
