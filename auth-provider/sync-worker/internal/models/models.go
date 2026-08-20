package models

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name                  string    `gorm:"type:varchar(255);not null"`
	ClientID              string    `gorm:"type:varchar(255);not null"`
	Status                string    `gorm:"type:varchar(50);not null"`
	LaunchURL             string    `gorm:"type:text"`
	LogoutNotificationURL string    `gorm:"type:text;not null"`
	CreatedAt             time.Time `gorm:"not null"`
	UpdatedAt             time.Time `gorm:"not null"`
}

func (Application) TableName() string {
	return "applications"
}

type Event struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	EventType        string     `gorm:"type:varchar(100);not null" json:"eventType"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	CentralSessionID *uuid.UUID `gorm:"type:uuid" json:"centralSessionId,omitempty"`
	ApplicationID    *uuid.UUID `gorm:"type:uuid" json:"applicationId,omitempty"`
	Payload          string     `gorm:"type:jsonb;not null" json:"payload"`
	Status           string     `gorm:"type:varchar(50);not null" json:"status"`
	CreatedAt        time.Time  `gorm:"not null" json:"createdAt"`
	PublishedAt      *time.Time `json:"publishedAt,omitempty"`
}

func (Event) TableName() string {
	return "events"
}

type EventDelivery struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID       uuid.UUID  `gorm:"type:uuid;not null"`
	ApplicationID uuid.UUID  `gorm:"type:uuid;not null"`
	Status        string     `gorm:"type:varchar(50);not null;default:'pending'"`
	AttemptCount  int        `gorm:"not null;default:0"`
	LastAttemptAt *time.Time
	NextRetryAt   *time.Time
	ProcessedAt   *time.Time
}

func (EventDelivery) TableName() string {
	return "event_deliveries"
}
