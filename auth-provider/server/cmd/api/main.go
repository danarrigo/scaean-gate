package main

import (
	"log"

	"github.com/danarrigo/scaean-gate/auth-provider/server/config"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/database"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/handler"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/middleware"
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
	auditRepo := repository.AuditRepository{DB: db}
	auditSvc := services.AuditService{Repo: auditRepo}
	appRepo := repository.AppRepository{DB: db}
	policyRepo := repository.PolicyRepository{DB: db}
	oauthRepo := repository.OAuthRepository{DB: db}

	authSvc := services.AuthService{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		AuditSvc:    auditSvc,
	}
	oauthSvc := services.OauthService{
		SessionRepo: sessionRepo,
		UserRepo:    userRepo,
		AppRepo:     appRepo,
		PolicyRepo:  policyRepo,
		OAuthRepo:   oauthRepo,
	}

	authHandler := handler.AuthHandler{
		AuthSvc: authSvc,
	}
	oauthHandler := handler.OauthHandler{
		OAuthSvc: oauthSvc,
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.LoggerMiddleware)
	r.POST("/login", authHandler.LoginHandler)
	r.POST("/logout", authHandler.Logout)
	r.POST("/change-password", authHandler.ChangePassword)
	r.GET("/profile", authHandler.ShowProfile)
	r.GET("/authorize", oauthHandler.Authorize)
	r.POST("/token", oauthHandler.Token)
	r.GET("/userinfo", oauthHandler.UserInfo)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
