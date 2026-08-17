// Package services implements application logic
package services

import (
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"github.com/google/uuid"
)

type AuditService struct {
	Repo repository.AuditRepository
}

func (s *AuditService) Log(eventType string, actorID, userID, appID, sessionID *uuid.UUID, result, ipAddress, reason string) {
	if actorID != nil && *actorID == uuid.Nil {
		actorID = nil
	}
	if userID != nil && *userID == uuid.Nil {
		userID = nil
	}
	if appID != nil && *appID == uuid.Nil {
		appID = nil
	}
	if sessionID != nil && *sessionID == uuid.Nil {
		sessionID = nil
	}

	metadata := "{}"
	if reason != "" {
		metadata = `{"reason":"` + reason + `"}`
	}

	_ = s.Repo.CreateAuditLog(&models.AuditLog{
		EventType:     eventType,
		ActorID:       actorID,
		UserID:        userID,
		ApplicationID: appID,
		SessionID:     sessionID,
		Result:        result,
		Metadata:      metadata,
		IPAddress:     ipAddress,
		CreatedAt:     time.Now(),
	})
}
