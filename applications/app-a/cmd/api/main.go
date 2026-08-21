package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/applications/app-a/config"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/handler"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/middleware"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/models"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/repository"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	if err := db.AutoMigrate(&models.LocalSession{}, &models.ProfileCache{}, &models.ProcessedEvent{}, &models.AuthActivity{}); err != nil {
		log.Fatalf("failed to auto-migrate database: %v", err)
	}

	repo := &repository.SessionRepository{DB: db}
	authSvc := services.NewAuthService(repo, cfg)
	authHandler := handler.NewAuthHandler(authSvc, cfg)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("RequestID", reqID)
		c.Header("X-Request-ID", reqID)

		origin := c.GetHeader("Origin")
		if origin == cfg.FrontendURL {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-Request-ID")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	liveHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	}
	readyHandler := func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "database"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
	r.GET("/health", readyHandler)
	r.GET("/health/live", liveHandler)
	r.GET("/health/ready", readyHandler)
	r.GET("/auth/login", authHandler.Login)
	r.GET("/auth/callback", authHandler.Callback)
	r.GET("/session-status", authHandler.SessionStatus)
	r.POST("/internal/logout", authHandler.InternalLogout)

	protected := r.Group("")
	protected.Use(middleware.AuthMiddleware(repo))
	{
		protected.GET("/me", authHandler.Me)
		protected.GET("/events", authHandler.GetEvents)
		protected.GET("/activity", authHandler.GetActivity)
		protected.POST("/logout", authHandler.Logout)
	}

	log.Printf("%s backend starting on port %s...", cfg.AppName, cfg.Port)
	if err := r.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
