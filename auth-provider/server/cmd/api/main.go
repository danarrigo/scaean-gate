package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/database"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "sso_user")
	dbPassword := getEnv("DB_PASSWORD", "sso_password")
	dbName := getEnv("DB_NAME", "sso_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	serverPort := getEnv("PORT", getEnv("SERVER_PORT", "8080"))

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("failed to seed columns: %v", err)
	}

	r := gin.Default()
	if err := r.Run(":" + serverPort); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
