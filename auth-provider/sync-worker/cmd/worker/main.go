package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/danarrigo/scaean-gate/auth-provider/sync-worker/config"
	"github.com/danarrigo/scaean-gate/auth-provider/sync-worker/internal/consumer"
	"github.com/danarrigo/scaean-gate/auth-provider/sync-worker/internal/dispatcher"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	disp := dispatcher.NewHTTPDispatcher(db, cfg.InternalAPISecret)

	brokerList := strings.Split(cfg.KafkaBrokers, ",")
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(brokerList...),
		kgo.ConsumeTopics(cfg.KafkaTopic),
		kgo.ConsumerGroup(cfg.KafkaGroupID),
	)
	if err != nil {
		log.Fatalf("failed to initialize kafka client: %v", err)
	}
	defer kafkaClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("shutting down sync worker...")
		cancel()
	}()

	kafkaCons := consumer.KafkaConsumer{
		Client:     kafkaClient,
		Dispatcher: disp,
	}

	log.Printf("sync worker started, listening on topic: %s", cfg.KafkaTopic)
	kafkaCons.Start(ctx)
}
