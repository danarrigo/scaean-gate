package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort       string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	SeedUserPassword string
	KafkaBrokers     string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = getEnvWithDefault("SERVER_PORT", "8080")
	}

	cfg := &Config{
		ServerPort:       serverPort,
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           getEnvWithDefault("DB_PORT", "5432"),
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           os.Getenv("DB_NAME"),
		DBSSLMode:        getEnvWithDefault("DB_SSLMODE", "disable"),
		SeedUserPassword: os.Getenv("SEED_USER_PASSWORD"),
		KafkaBrokers:     getEnvWithDefault("KAFKA_BROKERS", "localhost:9092"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
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
	if c.SeedUserPassword == "" {
		return fmt.Errorf("missing required environment variable: SEED_USER_PASSWORD")
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
