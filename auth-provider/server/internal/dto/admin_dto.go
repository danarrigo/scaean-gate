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
	GroupIDs []uuid.UUID `json:"groupIds"`
}

type UpdateUserRequest struct {
	Name     string      `json:"name" binding:"required"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password,omitempty"`
	GroupIDs []uuid.UUID `json:"groupIds"`
}

type UserDetailResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Status    string          `json:"status"`
	Groups    []GroupResponse `json:"groups"`
	CreatedAt time.Time       `json:"createdAt"`
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
	UserID uuid.UUID `json:"userId" binding:"required"`
}

type GroupResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GroupDetailResponse struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Users       []UserHeader `json:"users"`
	CreatedAt   time.Time    `json:"createdAt"`
}

type UserHeader struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Status string    `json:"status"`
}

type CreateAppRequest struct {
	Name                  string   `json:"name" binding:"required"`
	LaunchURL             string   `json:"launchUrl" binding:"required,url"`
	LogoutNotificationURL string   `json:"logoutNotificationUrl" binding:"required,url"`
	RedirectURIs          []string `json:"redirectUris" binding:"required,min=1"`
}

type UpdateAppRequest struct {
	Name                  string `json:"name" binding:"required"`
	LaunchURL             string `json:"launchUrl" binding:"required,url"`
	LogoutNotificationURL string `json:"logoutNotificationUrl" binding:"required,url"`
	Status                string `json:"status" binding:"required,oneof=active inactive"`
}

type AddRedirectURIRequest struct {
	URI string `json:"uri" binding:"required,url"`
}

type CreateAppResponse struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	ClientID              string    `json:"clientId"`
	ClientSecret          string    `json:"clientSecret"`
	Status                string    `json:"status"`
	LaunchURL             string    `json:"launchUrl"`
	LogoutNotificationURL string    `json:"logoutNotificationUrl"`
	RedirectURIs          []string  `json:"redirectUris"`
	CreatedAt             time.Time `json:"createdAt"`
}

type AppResponse struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	ClientID              string    `json:"clientId"`
	Status                string    `json:"status"`
	LaunchURL             string    `json:"launchUrl"`
	LogoutNotificationURL string    `json:"logoutNotificationUrl"`
	RedirectURIs          []string  `json:"redirectUris"`
	CreatedAt             time.Time `json:"createdAt"`
}

type CreatePolicyRequest struct {
	ApplicationID uuid.UUID `json:"applicationId" binding:"required"`
	GroupID       uuid.UUID `json:"groupId" binding:"required"`
	Effect        string    `json:"effect" binding:"required,oneof=allow deny"`
}

type PolicyResponse struct {
	ID              uuid.UUID `json:"id"`
	ApplicationID   uuid.UUID `json:"applicationId"`
	ApplicationName string    `json:"applicationName,omitempty"`
	GroupID         uuid.UUID `json:"groupId"`
	GroupName       string    `json:"groupName,omitempty"`
	Effect          string    `json:"effect"`
	CreatedAt       time.Time `json:"createdAt"`
}

type AuditLogResponse struct {
	ID            uuid.UUID  `json:"id"`
	EventType     string     `json:"eventType"`
	ActorID       *uuid.UUID `json:"actorId,omitempty"`
	UserID        *uuid.UUID `json:"userId,omitempty"`
	ApplicationID *uuid.UUID `json:"applicationId,omitempty"`
	SessionID     *uuid.UUID `json:"sessionId,omitempty"`
	Result        string     `json:"result"`
	Metadata      string     `json:"metadata,omitempty"`
	IPAddress     string     `json:"ipAddress,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type EventResponse struct {
	ID            uuid.UUID  `json:"id"`
	EventType     string     `json:"eventType"`
	UserID        uuid.UUID  `json:"userId"`
	ApplicationID *uuid.UUID `json:"applicationId,omitempty"`
	Payload       string     `json:"payload"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
}
