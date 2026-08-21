package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	AppName           string
	ClientID          string
	ClientSecret      string
	AuthProviderURL   string
	RedirectURI       string
	FrontendURL       string
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
		Port:              getEnvWithDefault("PORT", "8081"),
		AppName:           getEnvWithDefault("APP_NAME", "Apex"),
		ClientID:          getEnvWithDefault("CLIENT_ID", "app-a-client-id"),
		ClientSecret:      os.Getenv("CLIENT_SECRET"),
		AuthProviderURL:   getEnvWithDefault("AUTH_PROVIDER_URL", "http://localhost:8080"),
		RedirectURI:       getEnvWithDefault("REDIRECT_URI", "http://localhost:8081/auth/callback"),
		FrontendURL:       getEnvWithDefault("FRONTEND_URL", "http://localhost:4201"),
		DBHost:            os.Getenv("DB_HOST"),
		DBPort:            getEnvWithDefault("DB_PORT", "5432"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            os.Getenv("DB_NAME"),
		DBSSLMode:         getEnvWithDefault("DB_SSLMODE", "disable"),
		InternalAPISecret: getEnvWithDefault("INTERNAL_API_SECRET", "super-secret-internal-key"),
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
