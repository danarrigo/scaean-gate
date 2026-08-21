package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBrokers      string
	KafkaTopic        string
	KafkaGroupID      string
	KafkaDLQTopic     string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	InternalAPISecret string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		KafkaBrokers:      getEnvWithDefault("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:        getEnvWithDefault("KAFKA_TOPIC", "sso-session-events"),
		KafkaGroupID:      getEnvWithDefault("KAFKA_GROUP_ID", "sync-worker-group"),
		KafkaDLQTopic:     getEnvWithDefault("KAFKA_DLQ_TOPIC", "sso-session-events-dlq"),
		DBHost:            os.Getenv("DB_HOST"),
		DBPort:            getEnvWithDefault("DB_PORT", "5432"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            os.Getenv("DB_NAME"),
		DBSSLMode:         getEnvWithDefault("DB_SSLMODE", "disable"),
		InternalAPISecret: os.Getenv("INTERNAL_API_SECRET"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.InternalAPISecret == "" {
		return fmt.Errorf("missing required environment variable: INTERNAL_API_SECRET")
	}
	if c.DBHost == "" {
		return fmt.Errorf("missing required environment variable: DB_HOST")
	}
	if c.DBUser == "" {
		return fmt.Errorf("missing required environment variable: DB_USER")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("missing required environment variable: DB_PASSWORD")
	}
	if c.DBName == "" {
		return fmt.Errorf("missing required environment variable: DB_NAME")
	}
	return nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode)
}

func getEnvWithDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
