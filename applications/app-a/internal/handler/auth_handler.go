package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danarrigo/scaean-gate/applications/app-a/config"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/dto"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/models"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/applications/app-a/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthSvc *services.AuthService
	Cfg     *config.Config
}

func NewAuthHandler(authSvc *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		AuthSvc: authSvc,
		Cfg:     cfg,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	authURL, verifier, state, err := h.AuthSvc.GenerateLoginURL()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_verifier", verifier, 300, "/", "", false, true)
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Missing code or state parameter")
		return
	}

	savedState, err := c.Cookie("oauth_state")
	if err != nil || savedState != state {
		response.Error(c, http.StatusBadRequest, "INVALID_STATE", "OAuth state mismatch or expired")
		return
	}

	verifier, err := c.Cookie("oauth_verifier")
	if err != nil || verifier == "" {
		response.Error(c, http.StatusBadRequest, "INVALID_VERIFIER", "Missing PKCE verifier cookie")
		return
	}

	c.SetCookie("oauth_verifier", "", -1, "/", "", false, true)
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	rawLocalToken, err := h.AuthSvc.HandleCallback(c.Request.Context(), code, verifier)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("app_a_session", rawLocalToken, int(24*time.Hour.Seconds()), "/", "", false, true)

	c.Redirect(http.StatusFound, fmt.Sprintf("%s/dashboard", h.Cfg.FrontendURL))
}

func (h *AuthHandler) Me(c *gin.Context) {
	sessionVal, exists := c.Get("session")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	session := sessionVal.(*models.LocalSession)

	meResp, err := h.AuthSvc.GetMe(session)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, meResp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionVal, exists := c.Get("session")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}
	session := sessionVal.(*models.LocalSession)

	if err := h.AuthSvc.Logout(session.ID); err != nil {
		response.HandleError(c, err)
		return
	}

	c.SetCookie("app_a_session", "", -1, "/", "", false, true)
	response.JSON(c, http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) InternalLogout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid authorization header")
		return
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token != h.Cfg.InternalAPISecret {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid internal API secret")
		return
	}

	var payload dto.BackChannelLogoutPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid logout payload")
		return
	}

	if err := h.AuthSvc.HandleBackChannelLogout(payload); err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "success"})
}

func (h *AuthHandler) GetEvents(c *gin.Context) {
	events, err := h.AuthSvc.GetProcessedEvents()
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, events)
}
