// Package consumer
package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/danarrigo/scaean-gate/auth-provider/sync-worker/internal/dispatcher"
	"github.com/danarrigo/scaean-gate/auth-provider/sync-worker/internal/models"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaConsumer struct {
	Dispatcher *dispatcher.HTTPDispatcher
	Client     *kgo.Client
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			fetches := c.Client.PollFetches(ctx)

			if fetches.IsClientClosed() {
				return
			}

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					log.Printf("Fetch error : %v", err)
				}
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()

				var event models.Event
				if err := json.Unmarshal(record.Value, &event); err != nil {
					log.Printf("Failed to parse JSON : %v", event.ID)
					continue
				}

				if err := c.Dispatcher.Dispatch(ctx, event); err != nil {
					log.Printf("Failed to Dispatch %s : %v ", event.ID, err)
				}
			}
		}
	}
}
