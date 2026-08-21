// Package dispatcher contains back-channel logout delivery logic.
package dispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/sync-worker/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HTTPDispatcher struct {
	APIKey     string
	Client     *http.Client
	DB         *gorm.DB
	MaxRetries int
}

func NewHTTPDispatcher(db *gorm.DB, apiKey string) *HTTPDispatcher {
	return &HTTPDispatcher{
		APIKey:     apiKey,
		Client:     &http.Client{Timeout: 5 * time.Second},
		DB:         db,
		MaxRetries: 5,
	}
}

type Payload struct {
	EventID          uuid.UUID  `json:"event_id"`
	EventType        string     `json:"event_type"`
	UserID           uuid.UUID  `json:"user_id"`
	CentralSessionID *uuid.UUID `json:"central_session_id,omitempty"`
	Reason           string     `json:"reason"`
}

func (d *HTTPDispatcher) Dispatch(ctx context.Context, event models.Event) error {
	var targetApps []models.Application
	if event.EventType == "AccessPolicyChanged" && event.ApplicationID != nil {
		var app models.Application
		if err := d.DB.Where("id = ? AND status = 'active'", *event.ApplicationID).First(&app).Error; err != nil {
			return err
		}
		targetApps = append(targetApps, app)
	} else if err := d.DB.Where("status = ?", "active").Find(&targetApps).Error; err != nil {
		return err
	}

	var deliveryErrors []error
	for _, app := range targetApps {
		if err := d.dispatchToApp(ctx, event, app); err != nil {
			deliveryErrors = append(deliveryErrors, fmt.Errorf("%s: %w", app.Name, err))
		}
	}
	return errors.Join(deliveryErrors...)
}

func (d *HTTPDispatcher) dispatchToApp(parentCtx context.Context, event models.Event, app models.Application) error {
	delivery, err := d.getOrCreateDelivery(event.ID, app.ID)
	if err != nil {
		return err
	}
	if delivery.Status == "succeeded" {
		return nil
	}
	if delivery.AttemptCount >= d.MaxRetries {
		return fmt.Errorf("delivery exhausted after %d attempts: %s", delivery.AttemptCount, delivery.LastError)
	}

	payload, err := json.Marshal(Payload{
		EventID: event.ID, EventType: event.EventType, UserID: event.UserID,
		CentralSessionID: event.CentralSessionID, Reason: event.EventType,
	})
	if err != nil {
		return d.failDelivery(delivery, err)
	}

	for delivery.AttemptCount < d.MaxRetries {
		delivery.AttemptCount++
		now := time.Now()
		delivery.LastAttemptAt = &now
		delivery.NextRetryAt = nil
		delivery.Status = "processing"
		if err := d.DB.Save(delivery).Error; err != nil {
			return err
		}

		transient, requestErr := d.send(parentCtx, app.LogoutNotificationURL, payload)
		if requestErr == nil {
			delivery.Status = "succeeded"
			delivery.LastError = ""
			processedAt := time.Now()
			delivery.ProcessedAt = &processedAt
			return d.DB.Save(delivery).Error
		}

		delivery.LastError = requestErr.Error()
		if !transient || delivery.AttemptCount >= d.MaxRetries {
			delivery.Status = "failed"
			if err := d.DB.Save(delivery).Error; err != nil {
				return err
			}
			return requestErr
		}

		backoff := time.Duration(1<<(delivery.AttemptCount-1)) * time.Second
		nextRetry := time.Now().Add(backoff)
		delivery.Status = "retrying"
		delivery.NextRetryAt = &nextRetry
		if err := d.DB.Save(delivery).Error; err != nil {
			return err
		}

		timer := time.NewTimer(backoff)
		select {
		case <-parentCtx.Done():
			timer.Stop()
			return parentCtx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("delivery failed")
}

func (d *HTTPDispatcher) getOrCreateDelivery(eventID, applicationID uuid.UUID) (*models.EventDelivery, error) {
	var delivery models.EventDelivery
	err := d.DB.Where("event_id = ? AND application_id = ?", eventID, applicationID).First(&delivery).Error
	if err == nil {
		return &delivery, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	delivery = models.EventDelivery{
		EventID: eventID, ApplicationID: applicationID, Status: "pending",
	}
	if err := d.DB.Create(&delivery).Error; err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (d *HTTPDispatcher) send(parentCtx context.Context, targetURL string, payload []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	res, err := d.Client.Do(req)
	if err != nil {
		return true, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return false, nil
	}

	err = fmt.Errorf("back-channel logout returned HTTP %d", res.StatusCode)
	return res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500, err
}

func (d *HTTPDispatcher) failDelivery(delivery *models.EventDelivery, deliveryErr error) error {
	delivery.Status = "failed"
	delivery.LastError = deliveryErr.Error()
	if err := d.DB.Save(delivery).Error; err != nil {
		return err
	}
	return deliveryErr
}
