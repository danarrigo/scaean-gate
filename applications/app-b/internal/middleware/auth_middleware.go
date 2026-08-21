package middleware

import (
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/applications/app-b/internal/pkg/crypto"
	"github.com/danarrigo/scaean-gate/applications/app-b/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/applications/app-b/internal/repository"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(repo *repository.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, err := c.Cookie("app_b_session")
		if err != nil || rawToken == "" {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing local session cookie")
			c.Abort()
			return
		}

		tokenHash := crypto.HashSHA256(rawToken)
		session, err := repo.GetLocalSessionByTokenHash(tokenHash)
		if err != nil || session.Status != "active" || session.ExpiresAt.Before(time.Now()) {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or revoked local session")
			c.Abort()
			return
		}

		c.Set("session", session)
		c.Next()
	}
}
