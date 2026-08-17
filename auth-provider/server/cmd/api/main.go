package main

import (
	"log"

	"github.com/danarrigo/scaean-gate/auth-provider/server/config"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/database"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/handler"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
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

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	userRepo := repository.UserRepository{DB: db}
	sessionRepo := repository.SessionRepository{DB: db}
	authSvc := services.AuthService{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
	}
	userHandler := handler.UserHandler{
		AuthSvc: authSvc,
	}

	r := gin.Default()
	r.POST("/login", userHandler.LoginHandler)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
