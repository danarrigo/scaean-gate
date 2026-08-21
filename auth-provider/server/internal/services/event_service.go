package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type EventService struct {
	Client     *kgo.Client
	Topic      string
	OutboxRepo repository.OutboxRepository
}

func (s *EventService) RunOutboxPublisher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := s.publishPending(ctx); err != nil && ctx.Err() == nil {
			log.Printf("failed to publish outbox event: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *EventService) publishPending(ctx context.Context) error {
	events, err := s.OutboxRepo.ListPendingEvents(100)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := s.publishEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *EventService) publishEvent(parentCtx context.Context, event models.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	key := event.EventType
	if event.UserID != uuid.Nil {
		key = event.UserID.String()
	}

	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	record := &kgo.Record{Topic: s.Topic, Key: []byte(key), Value: payload}
	if err := s.Client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return err
	}
	return s.OutboxRepo.MarkPublished(event.ID)
}
