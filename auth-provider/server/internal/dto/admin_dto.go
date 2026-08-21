package dto

import (
	"time"

	"github.com/google/uuid"
)

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

type CreateUserRequest struct {
	Name     string      `json:"name" binding:"required"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=8"`
	GroupIDs []uuid.UUID `json:"group_ids"`
}

type UpdateUserRequest struct {
	Name     string      `json:"name" binding:"required"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password,omitempty"`
	GroupIDs []uuid.UUID `json:"group_ids"`
}

type UserDetailResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Status    string          `json:"status"`
	Groups    []GroupResponse `json:"groups"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type AssignUserGroupRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type GroupResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type GroupDetailResponse struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Users       []UserHeader `json:"users"`
	CreatedAt   time.Time    `json:"created_at"`
}

type UserHeader struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Status string    `json:"status"`
}

type CreateAppRequest struct {
	Name                  string   `json:"name" binding:"required"`
	LaunchURL             string   `json:"launch_url" binding:"required,url"`
	LogoutNotificationURL string   `json:"logout_notification_url" binding:"required,url"`
	RedirectURIs          []string `json:"redirect_uris" binding:"required,min=1"`
}

type UpdateAppRequest struct {
	Name                  string `json:"name" binding:"required"`
	LaunchURL             string `json:"launch_url" binding:"required,url"`
	LogoutNotificationURL string `json:"logout_notification_url" binding:"required,url"`
	Status                string `json:"status" binding:"required,oneof=active inactive"`
}

type AddRedirectURIRequest struct {
	URI string `json:"uri" binding:"required,url"`
}

type CreateAppResponse struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	ClientID              string    `json:"client_id"`
	ClientSecret          string    `json:"client_secret"`
	ClientSecretPrefix    string    `json:"client_secret_prefix"`
	Status                string    `json:"status"`
	LaunchURL             string    `json:"launch_url"`
	LogoutNotificationURL string    `json:"logout_notification_url"`
	RedirectURIs          []string  `json:"redirect_uris"`
	CreatedAt             time.Time `json:"created_at"`
}

type RedirectURIResponse struct {
	ID          uuid.UUID `json:"id"`
	RedirectURI string    `json:"redirect_uri"`
}

type AppResponse struct {
	ID                    uuid.UUID             `json:"id"`
	Name                  string                `json:"name"`
	ClientID              string                `json:"client_id"`
	ClientSecretPrefix    string                `json:"client_secret_prefix"`
	Status                string                `json:"status"`
	LaunchURL             string                `json:"launch_url"`
	LogoutNotificationURL string                `json:"logout_notification_url"`
	RedirectURIs          []string              `json:"redirect_uris"`
	RedirectURIItems      []RedirectURIResponse `json:"redirect_uri_items"`
	CreatedAt             time.Time             `json:"created_at"`
}

type CreatePolicyRequest struct {
	ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	GroupID       uuid.UUID `json:"group_id" binding:"required"`
	Effect        string    `json:"effect" binding:"required,oneof=allow deny"`
}

type UpdatePolicyRequest struct {
	ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	GroupID       uuid.UUID `json:"group_id" binding:"required"`
	Effect        string    `json:"effect" binding:"required,oneof=allow deny"`
}

type PolicyResponse struct {
	ID              uuid.UUID `json:"id"`
	ApplicationID   uuid.UUID `json:"application_id"`
	ApplicationName string    `json:"application_name,omitempty"`
	GroupID         uuid.UUID `json:"group_id"`
	GroupName       string    `json:"group_name,omitempty"`
	Effect          string    `json:"effect"`
	CreatedAt       time.Time `json:"created_at"`
}

type AuditLogResponse struct {
	ID            uuid.UUID  `json:"id"`
	EventType     string     `json:"event_type"`
	ActorID       *uuid.UUID `json:"actor_id,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	ApplicationID *uuid.UUID `json:"application_id,omitempty"`
	SessionID     *uuid.UUID `json:"session_id,omitempty"`
	Result        string     `json:"result"`
	Metadata      string     `json:"metadata,omitempty"`
	IPAddress     string     `json:"ip_address,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type EventDeliveryResponse struct {
	ID              uuid.UUID  `json:"id"`
	ApplicationID   uuid.UUID  `json:"application_id"`
	ApplicationName string     `json:"application_name"`
	Status          string     `json:"status"`
	AttemptCount    int        `json:"attempt_count"`
	NextRetryAt     *time.Time `json:"next_retry_at,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
}

type EventResponse struct {
	ID            uuid.UUID               `json:"id"`
	EventType     string                  `json:"event_type"`
	UserID        uuid.UUID               `json:"user_id"`
	ApplicationID *uuid.UUID              `json:"application_id,omitempty"`
	Status        string                  `json:"status"`
	CreatedAt     time.Time               `json:"created_at"`
	Deliveries    []EventDeliveryResponse `json:"deliveries"`
}
