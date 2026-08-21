package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort           string
	FrontendURL          string
	AllowedOrigins       []string
	CookieSecure         bool
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	SeedUserPassword     string
	SeedAppAClientSecret string
	SeedAppBClientSecret string
	AppALogoutURL        string
	AppBLogoutURL        string
	KafkaBrokers         string
	KafkaTopic           string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = getEnvWithDefault("SERVER_PORT", "8080")
	}

	cookieSecure, err := strconv.ParseBool(getEnvWithDefault("COOKIE_SECURE", "false"))
	if err != nil {
		return nil, fmt.Errorf("invalid COOKIE_SECURE: %w", err)
	}

	frontendURL := getEnvWithDefault("FRONTEND_URL", "http://localhost:4200")
	cfg := &Config{
		ServerPort:           serverPort,
		FrontendURL:          frontendURL,
		AllowedOrigins:       splitCSV(getEnvWithDefault("ALLOWED_ORIGINS", frontendURL)),
		CookieSecure:         cookieSecure,
		DBHost:               os.Getenv("DB_HOST"),
		DBPort:               getEnvWithDefault("DB_PORT", "5432"),
		DBUser:               os.Getenv("DB_USER"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               os.Getenv("DB_NAME"),
		DBSSLMode:            getEnvWithDefault("DB_SSLMODE", "disable"),
		SeedUserPassword:     os.Getenv("SEED_USER_PASSWORD"),
		SeedAppAClientSecret: os.Getenv("APP_A_CLIENT_SECRET"),
		SeedAppBClientSecret: os.Getenv("APP_B_CLIENT_SECRET"),
		AppALogoutURL:        getEnvWithDefault("APP_A_LOGOUT_URL", "http://localhost:8081/internal/logout"),
		AppBLogoutURL:        getEnvWithDefault("APP_B_LOGOUT_URL", "http://localhost:8082/internal/logout"),
		KafkaBrokers:         getEnvWithDefault("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:           getEnvWithDefault("KAFKA_TOPIC", "sso-session-events"),
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
	if c.SeedAppAClientSecret == "" {
		return fmt.Errorf("missing required environment variable: APP_A_CLIENT_SECRET")
	}
	if c.SeedAppBClientSecret == "" {
		return fmt.Errorf("missing required environment variable: APP_B_CLIENT_SECRET")
	}
	if len(c.AllowedOrigins) == 0 {
		return fmt.Errorf("ALLOWED_ORIGINS must contain at least one origin")
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

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
