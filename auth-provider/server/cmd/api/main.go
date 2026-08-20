package main

import (
	"log"

	"strings"

	"github.com/danarrigo/scaean-gate/auth-provider/server/config"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/database"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/handler"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/middleware"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
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

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	userRepo := repository.UserRepository{DB: db}
	sessionRepo := repository.SessionRepository{DB: db}
	groupRepo := repository.GroupRepository{DB: db}
	appRepo := repository.AppRepository{DB: db}
	policyRepo := repository.PolicyRepository{DB: db}
	oauthRepo := repository.OAuthRepository{DB: db}
	auditRepo := repository.AuditRepository{DB: db}

	brokerList := strings.Split(cfg.KafkaBrokers, ",")
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(brokerList...),
	)
	if err != nil {
		log.Printf("failed to initialize kafka client: %v", err)
	}
	defer func() {
		if kafkaClient != nil {
			kafkaClient.Close()
		}
	}()

	eventSvc := &services.EventService{
		Client:    kafkaClient,
		Topic:     "sso-session-events",
		AuditRepo: auditRepo,
	}

	auditSvc := services.AuditService{Repo: auditRepo}
	authSvc := services.AuthService{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		AuditSvc:    auditSvc,
		EventSvc:    eventSvc,
	}
	oauthSvc := services.OauthService{
		SessionRepo: sessionRepo,
		UserRepo:    userRepo,
		AppRepo:     appRepo,
		PolicyRepo:  policyRepo,
		OAuthRepo:   oauthRepo,
	}
	adminSvc := services.AdminService{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		GroupRepo:   groupRepo,
		AppRepo:     appRepo,
		PolicyRepo:  policyRepo,
		AuditRepo:   auditRepo,
		EventSvc:    eventSvc,
	}

	authHandler := handler.AuthHandler{
		AuthSvc: authSvc,
	}
	oauthHandler := handler.OauthHandler{
		OAuthSvc: oauthSvc,
	}
	userHandler := handler.UserHandler{
		AdminSvc: adminSvc,
	}
	groupHandler := handler.GroupHandler{
		AdminSvc: adminSvc,
	}
	appHandler := handler.AppHandler{
		AdminSvc: adminSvc,
	}
	policyHandler := handler.PolicyHandler{
		AdminSvc: adminSvc,
	}
	auditHandler := handler.AuditHandler{
		AdminSvc: adminSvc,
	}
	authMiddleware := middleware.AuthMiddlewareHandler{
		AuthSvc: adminSvc,
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

	admin := r.Group("/admin")
	admin.Use(authMiddleware.AuthMiddleWare)
	{
		admin.GET("/users", userHandler.ListUsers)
		admin.POST("/users", userHandler.CreateUser)
		admin.GET("/users/:id", userHandler.GetUser)
		admin.PUT("/users/:id", userHandler.UpdateUser)
		admin.DELETE("/users/:id", userHandler.DeleteUser)
		admin.PATCH("/users/:id/status", userHandler.UpdateUserStatus)

		admin.GET("/groups", groupHandler.ListGroups)
		admin.POST("/groups", groupHandler.CreateGroup)
		admin.GET("/groups/:id", groupHandler.GetGroup)
		admin.PUT("/groups/:id", groupHandler.UpdateGroup)
		admin.DELETE("/groups/:id", groupHandler.DeleteGroup)
		admin.POST("/groups/:id/users", groupHandler.AssignUser)
		admin.DELETE("/groups/:id/users/:user_id", groupHandler.UnassignUser)

		admin.GET("/apps", appHandler.ListApps)
		admin.POST("/apps", appHandler.CreateApp)
		admin.GET("/apps/:id", appHandler.GetApp)
		admin.PUT("/apps/:id", appHandler.UpdateApp)
		admin.DELETE("/apps/:id", appHandler.DeleteApp)
		admin.POST("/apps/:id/redirect-uris", appHandler.AddRedirectURI)
		admin.DELETE("/apps/:id/redirect-uris/:uri_id", appHandler.DeleteRedirectURI)

		admin.GET("/policies", policyHandler.ListPolicies)
		admin.POST("/policies", policyHandler.CreatePolicy)
		admin.DELETE("/policies/:id", policyHandler.DeletePolicy)

		admin.GET("/audit-logs", auditHandler.ListAuditLogs)
		admin.GET("/events", auditHandler.ListEvents)
	}

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
