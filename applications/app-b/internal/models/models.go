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
	CentralSessionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"centralSessionId"`
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
	EventID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"eventId"`
	EventType   string    `gorm:"type:varchar(100);not null" json:"eventType"`
	ProcessedAt time.Time `gorm:"not null" json:"processedAt"`
	Result      string    `gorm:"type:varchar(50);not null" json:"result"`
}
