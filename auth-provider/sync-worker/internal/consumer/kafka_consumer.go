// Package consumer contains logic for consuming Kafka records.
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
	DLQTopic   string
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	for ctx.Err() == nil {
		fetches := c.Client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}
		for _, fetchErr := range fetches.Errors() {
			log.Printf("Kafka fetch error: %v", fetchErr)
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			if err := c.processRecord(ctx, record); err != nil {
				log.Printf("Failed to process Kafka record at %s[%d] offset %d: %v", record.Topic, record.Partition, record.Offset, err)
			}
		}
	}
}

func (c *KafkaConsumer) processRecord(ctx context.Context, record *kgo.Record) error {
	var event models.Event
	if err := json.Unmarshal(record.Value, &event); err != nil {
		if dlqErr := c.publishDLQ(ctx, record, err); dlqErr != nil {
			return dlqErr
		}
		return c.Client.CommitRecords(ctx, record)
	}

	if err := c.Dispatcher.Dispatch(ctx, event); err != nil {
		if dlqErr := c.publishDLQ(ctx, record, err); dlqErr != nil {
			return dlqErr
		}
	}
	return c.Client.CommitRecords(ctx, record)
}

func (c *KafkaConsumer) publishDLQ(ctx context.Context, source *kgo.Record, processingErr error) error {
	record := &kgo.Record{
		Topic: c.DLQTopic,
		Key:   source.Key,
		Value: source.Value,
		Headers: []kgo.RecordHeader{
			{Key: "source-topic", Value: []byte(source.Topic)},
			{Key: "processing-error", Value: []byte(processingErr.Error())},
		},
	}
	return c.Client.ProduceSync(ctx, record).FirstErr()
}
