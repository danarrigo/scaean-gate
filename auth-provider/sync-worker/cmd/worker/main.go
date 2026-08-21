package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to access database pool: %v", err)
	}
	defer sqlDB.Close()

	disp := dispatcher.NewHTTPDispatcher(db, cfg.InternalAPISecret)

	brokerList := strings.Split(cfg.KafkaBrokers, ",")
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(brokerList...),
		kgo.ConsumeTopics(cfg.KafkaTopic),
		kgo.ConsumerGroup(cfg.KafkaGroupID),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		log.Fatalf("failed to initialize kafka client: %v", err)
	}
	defer kafkaClient.Close()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	if err := kafkaClient.Ping(pingCtx); err != nil {
		cancelPing()
		log.Fatalf("failed to connect to kafka: %v", err)
	}
	cancelPing()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"alive"}`))
	})
	healthMux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		readyCtx, readyCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer readyCancel()
		w.Header().Set("Content-Type", "application/json")
		if err := sqlDB.PingContext(readyCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready","dependency":"database"}`))
			return
		}
		if err := kafkaClient.Ping(readyCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready","dependency":"broker"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})
	healthServer := &http.Server{Addr: ":" + cfg.HealthPort, Handler: healthMux, ReadHeaderTimeout: 2 * time.Second}
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("health server failed: %v", err)
			cancel()
		}
	}()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		_ = healthServer.Shutdown(shutdownCtx)
	}()

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
		DLQTopic:   cfg.KafkaDLQTopic,
	}

	log.Printf("sync worker started, listening on topic: %s", cfg.KafkaTopic)
	kafkaCons.Start(ctx)
}
