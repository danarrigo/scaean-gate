// Package dispatcher contains logic so the sync-worker can send notification to the apps that use the SSO
package dispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
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
		APIKey: apiKey,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
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
	targetApps := make([]models.Application, 0)

	if event.EventType == "AccessPolicyChanged" && event.ApplicationID != nil {
		var app models.Application
		if err := d.DB.Where("id = ? AND status = 'active'", *event.ApplicationID).First(&app).Error; err != nil {
			return err
		}
		targetApps = append(targetApps, app)
	} else {
		if err := d.DB.Where("status = ?", "active").Find(&targetApps).Error; err != nil {
			return err
		}
	}

	for _, app := range targetApps {
		d.DispatchToApp(ctx, event, app)
	}
	return nil
}

func (d *HTTPDispatcher) DispatchToApp(parentCtx context.Context, event models.Event, app models.Application) {
	var delivery models.EventDelivery
	if err := d.DB.Where("event_id = ? AND application_id = ?", event.ID, app.ID).First(&delivery).Error; err != nil {
		now := time.Now()
		delivery = models.EventDelivery{
			EventID:       event.ID,
			ApplicationID: app.ID,
			Status:        "processing",
			AttemptCount:  1,
			LastAttemptAt: &now,
		}
		if err := d.DB.Create(&delivery).Error; err != nil {
			return
		}
	} else {
		if delivery.Status == "succeeded" {
			return
		}
		now := time.Now()
		delivery.LastAttemptAt = &now
		delivery.Status = "processing"
		if err := d.DB.Save(&delivery).Error; err != nil {
			return
		}
	}

	bodyPayload := Payload{
		EventID:          event.ID,
		EventType:        event.EventType,
		UserID:           event.UserID,
		CentralSessionID: event.CentralSessionID,
		Reason:           event.EventType,
	}

	payload, err := json.Marshal(bodyPayload)
	if err != nil {
		_ = d.markDeliveryAsFailed(delivery.ID)
		return
	}

	for attempt := 1; attempt <= d.MaxRetries; attempt++ {
		delivery.AttemptCount = attempt
		now := time.Now()
		delivery.LastAttemptAt = &now

		ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, app.LogoutNotificationURL, bytes.NewBuffer(payload))
		if err != nil {
			cancel()
			_ = d.markDeliveryAsFailed(delivery.ID)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+d.APIKey)

		res, err := d.Client.Do(req)
		cancel()

		if err == nil && res != nil && res.StatusCode >= 200 && res.StatusCode < 300 {
			_ = res.Body.Close()
			delivery.Status = "succeeded"
			processedAt := time.Now()
			delivery.ProcessedAt = &processedAt
			d.DB.Save(&delivery)
			return
		}

		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}

		_ = d.handleDeliveryError(delivery.ID)

		if attempt < d.MaxRetries {
			interval := math.Pow(2, float64(attempt)) * 2
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

func (d *HTTPDispatcher) markDeliveryAsFailed(deliveryID uuid.UUID) error {
	return d.DB.Model(&models.EventDelivery{}).Where("id = ?", deliveryID).Update("status", "failed").Error
}

func (d *HTTPDispatcher) handleDeliveryError(deliveryID uuid.UUID) error {
	var delivery models.EventDelivery
	if err := d.DB.Where("id = ?", deliveryID).First(&delivery).Error; err != nil {
		return err
	}
	if delivery.AttemptCount >= d.MaxRetries {
		delivery.Status = "failed"
	} else {
		delivery.Status = "retrying"
		interval := math.Pow(2, float64(delivery.AttemptCount)) * 2
		nextRetry := time.Now().Add(time.Duration(interval) * time.Second)
		delivery.NextRetryAt = &nextRetry
	}

	return d.DB.Save(&delivery).Error
}
