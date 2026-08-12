// Package models contains database schema
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string         `gorm:"type:varchar(255);not null"`
	Email        string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_user_email_deleted"`
	PasswordHash string         `gorm:"type:varchar(255);not null"`
	Status       string         `gorm:"type:varchar(50);not null;default:'active'"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index;uniqueIndex:idx_user_email_deleted"`

	Groups []Group `gorm:"many2many:user_groups;"`
}

type Group struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Users []User `gorm:"many2many:user_groups;"`
}

type UserGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_group"`
	GroupID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_group"`
	CreatedAt time.Time `gorm:"not null"`

	User  User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Group Group `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Application struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                  string    `gorm:"type:varchar(255);not null"`
	ClientID              string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ClientSecretHash      string    `gorm:"type:varchar(255)"`
	Status                string    `gorm:"type:varchar(50);not null;default:'active'"`
	LaunchURL             string    `gorm:"type:text"`
	LogoutNotificationURL string    `gorm:"type:text;not null"`
	CreatedAt             time.Time `gorm:"not null"`
	UpdatedAt             time.Time `gorm:"not null"`

	RedirectURIs []ApplicationRedirectURI `gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ApplicationRedirectURI struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ApplicationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_app_redirect"`
	RedirectURI   string    `gorm:"type:text;not null;uniqueIndex:idx_app_redirect"`
	CreatedAt     time.Time `gorm:"not null"`

	Application Application `gorm:"foreignKey:ApplicationID"`
}

type ApplicationGroupPolicy struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ApplicationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_app_group_effect"`
	GroupID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_app_group_effect"`
	Effect        string    `gorm:"type:varchar(50);not null;default:'allow';uniqueIndex:idx_app_group_effect"`
	CreatedAt     time.Time `gorm:"not null"`

	Application Application `gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Group       Group       `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type SSOSession struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index"`
	SessionTokenHash string    `gorm:"type:varchar(255);not null;index"`
	Status           string    `gorm:"type:varchar(50);not null;default:'active'"`
	CreatedAt        time.Time `gorm:"not null"`
	ExpiresAt        time.Time `gorm:"not null"`
	RevokedAt        *time.Time
	RevokeReason     string `gorm:"type:varchar(255)"`
	IPAddress        string `gorm:"type:varchar(45)"`
	UserAgent        string `gorm:"type:text"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AuthorizationCode struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CodeHash            string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	UserID              uuid.UUID `gorm:"type:uuid;not null"`
	ApplicationID       uuid.UUID `gorm:"type:uuid;not null"`
	SSOSessionID        uuid.UUID `gorm:"type:uuid;not null"`
	RedirectURI         string    `gorm:"type:text;not null"`
	CodeChallenge       string    `gorm:"type:varchar(255);not null"`
	CodeChallengeMethod string    `gorm:"type:varchar(50);not null;default:'S256'"`
	CreatedAt           time.Time `gorm:"not null"`
	ExpiresAt           time.Time `gorm:"not null"`
	UsedAt              *time.Time

	User        User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Application Application `gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SSOSession  SSOSession  `gorm:"foreignKey:SSOSessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AccessToken struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TokenHash     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	ApplicationID uuid.UUID `gorm:"type:uuid;not null"`
	SSOSessionID  uuid.UUID `gorm:"type:uuid;not null"`
	Status        string    `gorm:"type:varchar(50);not null;default:'active'"`
	IssuedAt      time.Time `gorm:"not null"`
	ExpiresAt     time.Time `gorm:"not null"`
	RevokedAt     *time.Time

	User        User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Application Application `gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SSOSession  SSOSession  `gorm:"foreignKey:SSOSessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AuditLog struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventType     string     `gorm:"type:varchar(100);not null;index"`
	ActorID       *uuid.UUID `gorm:"type:uuid"`
	UserID        *uuid.UUID `gorm:"type:uuid"`
	ApplicationID *uuid.UUID `gorm:"type:uuid"`
	SessionID     *uuid.UUID `gorm:"type:uuid"`
	Result        string     `gorm:"type:varchar(50);not null"`
	Metadata      string     `gorm:"type:jsonb"`
	IPAddress     string     `gorm:"type:varchar(45)"`
	CreatedAt     time.Time  `gorm:"not null"`
}

type Event struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventType        string     `gorm:"type:varchar(100);not null"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null"`
	CentralSessionID *uuid.UUID `gorm:"type:uuid"`
	ApplicationID    *uuid.UUID `gorm:"type:uuid"`
	Payload          string     `gorm:"type:jsonb;not null"`
	Status           string     `gorm:"type:varchar(50);not null;default:'pending'"`
	CreatedAt        time.Time  `gorm:"not null"`
	PublishedAt      *time.Time
}

type EventDelivery struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_app"`
	ApplicationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_app"`
	Status        string    `gorm:"type:varchar(50);not null;default:'pending'"`
	AttemptCount  int       `gorm:"not null;default:0"`
	LastAttemptAt *time.Time
	NextRetryAt   *time.Time
	ProcessedAt   *time.Time
	LastError     string `gorm:"type:text"`

	Event       Event       `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Application Application `gorm:"foreignKey:ApplicationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
