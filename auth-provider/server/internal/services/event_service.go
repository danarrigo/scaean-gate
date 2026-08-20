package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type EventService struct {
	Client    *kgo.Client
	Topic     string
	AuditRepo repository.AuditRepository
}

func (s *EventService) PublishEvent(event models.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	key := event.EventType
	if event.UserID != uuid.Nil {
		key = event.UserID.String()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record := &kgo.Record{
		Topic: s.Topic,
		Key:   []byte(key),
		Value: payload,
	}

	if err := s.Client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return err
	}

	if err := s.AuditRepo.MarkEventAsFinished(event.ID); err != nil {
		return err
	}

	return nil
}
