package dto

import (
	"time"

	"github.com/google/uuid"
)

type TokenResponse struct {
	AccessToken      string     `json:"access_token"`
	TokenType        string     `json:"token_type"`
	ExpiresIn        int        `json:"expires_in"`
	UserID           uuid.UUID  `json:"user_id,omitempty"`
	CentralSessionID *uuid.UUID `json:"central_session_id,omitempty"`
}

type UserInfoResponse struct {
	Sub    string   `json:"sub"`
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}

type BackChannelLogoutPayload struct {
	EventID          uuid.UUID  `json:"eventId"`
	EventType        string     `json:"eventType"`
	UserID           uuid.UUID  `json:"userId"`
	CentralSessionID *uuid.UUID `json:"centralSessionId,omitempty"`
	Reason           string     `json:"reason"`
}

type UserProfileDTO struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Groups []string  `json:"groups"`
}

type SessionDTO struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type MeResponse struct {
	AppName string         `json:"appName"`
	User    UserProfileDTO `json:"user"`
	Session SessionDTO     `json:"session"`
}
