// Package models contains the database structure
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StringArray []string

func (a *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, a)
}

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

type LocalSession struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SessionTokenHash string     `gorm:"type:varchar(255);not null;index" json:"sessionTokenHash"`
	ExternalUserID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"externalUserId"`
	CentralSessionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"central_session_id"`
	OAuthFlowID      string     `gorm:"column:oauth_flow_id;type:varchar(255);not null;index" json:"oauth_flow_id"`
	Status           string     `gorm:"type:varchar(50);not null" json:"status"`
	CreatedAt        time.Time  `gorm:"not null" json:"createdAt"`
	ExpiresAt        time.Time  `gorm:"not null" json:"expiresAt"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
	RevokeReason     *string    `gorm:"type:varchar(255)" json:"revokeReason,omitempty"`
}

type ProfileCache struct {
	ExternalUserID uuid.UUID   `gorm:"type:uuid;primaryKey" json:"externalUserId"`
	Name           string      `gorm:"type:varchar(255);not null" json:"name"`
	Email          string      `gorm:"type:varchar(255);not null" json:"email"`
	Groups         StringArray `gorm:"type:jsonb" json:"groups"`
	SyncedAt       time.Time   `gorm:"not null" json:"syncedAt"`
	CreatedAt      time.Time   `gorm:"not null" json:"createdAt"`
	UpdatedAt      time.Time   `gorm:"not null" json:"updatedAt"`
}

type ProcessedEvent struct {
	EventID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"event_id"`
	EventType   string    `gorm:"type:varchar(100);not null" json:"event_type"`
	ProcessedAt time.Time `gorm:"not null" json:"processed_at"`
	Result      string    `gorm:"type:varchar(50);not null" json:"result"`
}

type AuthActivity struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OAuthFlowID    string     `gorm:"column:oauth_flow_id;type:varchar(255);not null;index" json:"oauth_flow_id"`
	LocalSessionID *uuid.UUID `gorm:"type:uuid;index" json:"local_session_id,omitempty"`
	EventType      string     `gorm:"type:varchar(100);not null" json:"event_type"`
	Result         string     `gorm:"type:varchar(50);not null" json:"result"`
	CreatedAt      time.Time  `gorm:"not null" json:"created_at"`
}
