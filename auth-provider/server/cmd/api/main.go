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
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to access database pool: %v", err)
	}
	defer sqlDB.Close()

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := database.Seed(db, cfg.SeedUserPassword, cfg.SeedAppAClientSecret, cfg.SeedAppBClientSecret, cfg.AppALogoutURL, cfg.AppBLogoutURL); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	userRepo := repository.UserRepository{DB: db}
	sessionRepo := repository.SessionRepository{DB: db}
	groupRepo := repository.GroupRepository{DB: db}
	appRepo := repository.AppRepository{DB: db}
	policyRepo := repository.PolicyRepository{DB: db}
	oauthRepo := repository.OAuthRepository{DB: db}
	auditRepo := repository.AuditRepository{DB: db}
	outboxRepo := repository.OutboxRepository{DB: db}

	brokerList := strings.Split(cfg.KafkaBrokers, ",")
	kafkaClient, err := kgo.NewClient(kgo.SeedBrokers(brokerList...))
	if err != nil {
		log.Fatalf("failed to initialize kafka client: %v", err)
	}
	defer kafkaClient.Close()

	kafkaCtx, cancelKafkaPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelKafkaPing()
	if err := kafkaClient.Ping(kafkaCtx); err != nil {
		log.Fatalf("failed to connect to kafka: %v", err)
	}

	eventSvc := &services.EventService{
		Client:     kafkaClient,
		Topic:      cfg.KafkaTopic,
		OutboxRepo: outboxRepo,
	}
	outboxCtx, cancelOutbox := context.WithCancel(context.Background())
	defer cancelOutbox()
	go eventSvc.RunOutboxPublisher(outboxCtx, time.Second)

	auditSvc := services.AuditService{Repo: auditRepo}
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
		AuditRepo:   auditRepo,
	}
	adminSvc := services.AdminService{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		GroupRepo:   groupRepo,
		AppRepo:     appRepo,
		PolicyRepo:  policyRepo,
		AuditRepo:   auditRepo,
	}

	authHandler := handler.AuthHandler{
		AuthSvc: authSvc,
		Cfg:     cfg,
	}
	oauthHandler := handler.OauthHandler{
		OAuthSvc: oauthSvc,
		Cfg:      cfg,
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
	r.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	r.Use(middleware.LoggerMiddleware)

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
		if err := kafkaClient.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependency": "broker"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
	r.GET("/health", readyHandler)
	r.GET("/health/live", liveHandler)
	r.GET("/health/ready", readyHandler)
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
		admin.GET("/policies/:id", policyHandler.GetPolicy)
		admin.PUT("/policies/:id", policyHandler.UpdatePolicy)
		admin.DELETE("/policies/:id", policyHandler.DeletePolicy)

		admin.GET("/audit-logs", auditHandler.ListAuditLogs)
		admin.GET("/events", auditHandler.ListEvents)
	}

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	select {
	case <-signalCtx.Done():
		log.Printf("shutting down auth provider...")
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
		}
		return
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown timed out: %v", err)
	}
	cancelOutbox()
}
