// Package handler consists of handlers
package handler

import (
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/config"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthSvc services.AuthService
	Cfg     *config.Config
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var body dto.LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	result, err := h.AuthSvc.Login(body.Email, body.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("sso_session", result.RawToken, int(24*time.Hour.Seconds()), "/", "", h.Cfg.CookieSecure, true)

	response.JSON(c, http.StatusOK, dto.LoginResponse{
		Message: "Login Successful",
		User: dto.UserData{
			ID:    result.User.ID,
			Name:  result.User.Name,
			Email: result.User.Email,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	cookie, err := c.Cookie("sso_session")
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "No active session")
		return
	}
	if err := h.AuthSvc.Logout(cookie, c.ClientIP()); err != nil {
		response.HandleError(c, err)
		return
	}

	c.SetCookie("sso_session", "", -1, "/", "", h.Cfg.CookieSecure, true)
	response.JSON(c, http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	cookie, err := c.Cookie("sso_session")
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "No active session")
		return
	}

	var body dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.AuthSvc.ChangePassword(cookie, body, c.ClientIP()); err != nil {
		response.HandleError(c, err)
		return
	}

	c.SetCookie("sso_session", "", -1, "/", "", h.Cfg.CookieSecure, true)
	response.JSON(c, http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *AuthHandler) ShowProfile(c *gin.Context) {
	cookie, err := c.Cookie("sso_session")
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "No active session")
		return
	}

	profile, err := h.AuthSvc.GetProfile(cookie)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, dto.ProfileResponse{User: *profile})
}
