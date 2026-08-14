// Package handler consists of handlers
package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body dto.LoginRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("INVALID_REQUEST", err.Error()))
			return
		}
		var user models.User
		if err := db.Where("email = ?", body.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("INVALID_CREDENTIALS", "Invalid email or password"))
			return
		}

		if user.Status != "active" {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("USER_INACTIVE", "User account is inactive"))
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("INVALID_CREDENTIALS", "Invalid email or password"))
			return
		}

		buf := make([]byte, 32)
		_, err := rand.Read(buf)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("INTERNAL_ERROR", "Failed to generate session token"))
			return
		}
		rawToken := hex.EncodeToString(buf)
		hashedBytes := sha256.Sum256([]byte(rawToken))
		hashedString := hex.EncodeToString(hashedBytes[:])

		SSOSession := models.SSOSession{
			UserID:           user.ID,
			SessionTokenHash: hashedString,
			Status:           "active",
			ExpiresAt:        time.Now().Add(24 * time.Hour),
			IPAddress:        c.ClientIP(),
			UserAgent:        c.Request.UserAgent(),
		}
		if err := db.Create(&SSOSession).Error; err != nil {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("INTERNAL_ERROR", "Failed to create session"))
			return
		}

		c.SetCookie("sso_session", rawToken, int(24*time.Hour.Seconds()), "/", "", false, true)

		response := dto.LoginResponse{
			Message: "Login Successful",
			User: dto.UserData{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			},
		}
		c.JSON(http.StatusOK, response)
	}
}

func Logout(c *gin.Context) {
}

func ChangePassword(c *gin.Context) {
}

func ShowProfile(c *gin.Context) {
}
